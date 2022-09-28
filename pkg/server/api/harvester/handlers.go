package harvester

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (s *echoServer) syncBundleHandler(ctx echo.Context) error {
	logger := ctx.Logger()

	// auth
	token := common.Member{
		ID:          uuid.New(),
		TrustDomain: spiffeid.RequireTrustDomainFromString("spiffe://td2"),
		// ...
	}
	logger.Info("Receiving sync request")

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		logger.Error("failed to read body: ", err)
		return err
	}
	receivedHarvesterState := common.SyncBundleBody{}
	err = json.Unmarshal(body, &receivedHarvesterState)
	if err != nil {
		logger.Error("failed to unmarshal state: ", err)
		return err
	}

	// get relationships for that member
	relationships, err := s.DataStore.GetRelationship(context.TODO(), token.TrustDomain.String())
	if err != nil {
		logger.Error("failed to fetch relationships: ", err)
		return err
	}

	response := common.SyncBundleResponse{}

	currentState, err := s.calculateBundleState(relationships, token.TrustDomain.String())
	if err != nil {
		logger.Error("failed calculating bundle state: ", err)
		return err
	}

	response.State = currentState
	response.Update = s.calculateBundleSync(receivedHarvesterState.FederatesWith, relationships, token.ID)

	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("failed to marshal response: ", err)
		return err
	}

	_, err = ctx.Response().Write(responseBytes)
	if err != nil {
		logger.Error("failed to write response: ", err)
		return err
	}

	return nil
}

func (s *echoServer) postBundleHandler(ctx echo.Context) error {
	logger := ctx.Logger()
	logger.Info("Receiving post bundle request")

	// auth
	token := common.Member{
		TrustDomain: spiffeid.RequireTrustDomainFromString("spiffe://td2"),
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		logger.Error("failed to read body: ", err)
		return err
	}

	receivedHarvesterState := common.PostBundleBody{}
	err = json.Unmarshal(body, &receivedHarvesterState)
	if err != nil {
		logger.Error("failed to unmarshal state: ", err)
		return err
	}

	// fetch harvester state from db
	currentState, err := s.DataStore.GetMember(context.TODO(), token.TrustDomain.String())
	if err != nil {
		logger.Error("failed to fetch member: ", err)
		return err
	}

	// update internal state
	if !bytes.Equal(receivedHarvesterState.TrustBundleHash, currentState.TrustBundleHash) {
		_, err := s.DataStore.UpdateMember(context.TODO(), token.TrustDomain.String(), &common.Member{
			TrustBundle:     receivedHarvesterState.TrustBundle,
			TrustBundleHash: receivedHarvesterState.TrustBundleHash,
		})
		if err != nil {
			logger.Error("failed to update member: ", err)
			return err
		}
		logger.Infof("Trust domain %s has been successfully updated", receivedHarvesterState.TrustDomain)
	}

	return nil
}

func (s *echoServer) onboardHandler(c echo.Context) error {
	header := c.Request().Header.Get("Authorization")
	err := s.validateToken(header, c)
	if err != nil {
		return c.String(400, "Invalid token")
	}

	c.Logger().Info("Harvester connected")
	return nil
}

func (s *echoServer) validateToken(header string, c echo.Context) error {
	var splitToken = strings.Split(header, "Bearer ")
	if len(splitToken) != 2 {
		return errors.New("invalid token")
	}
	token := splitToken[1]

	t, err := s.DataStore.FetchAccessToken(context.TODO(), token)
	if err != nil {
		c.Logger().Errorf("Invalid Token: %s\n", token)
		return err
	}

	c.Logger().Infof("Token valid for trust domain: %s\n", t.Member.TrustDomain)

	return nil
}

func (s *echoServer) calculateBundleSync(
	received common.FederationState,
	current []*datastore.Relationship,
	receivedID uuid.UUID) common.FederationState {
	// iterate over all found relationships, and for each, iterate over all received relationships.
	// if we can find the corresponding relationship, and their state match, continue.
	// if their state differ or if we cannot find a match, append an update request
	response := make(common.FederationState)
	for _, r := range current {
		found := false
		for receivedTD, receivedMember := range received {
			// uni-directional relationships will make this cleaner
			if r.MemberA.TrustDomain == receivedTD {
				if !bytes.Equal(receivedMember.TrustBundleHash, r.MemberA.TrustBundleHash) {
					// harvester trust bundle is outdated
					response[r.MemberA.TrustDomain] = common.MemberState{
						TrustDomain:     r.MemberA.TrustDomain,
						TrustBundle:     r.MemberA.TrustBundle,
						TrustBundleHash: r.MemberA.TrustBundleHash,
					}
				}
				found = true
			}

			if r.MemberB.TrustDomain == receivedTD {
				if !bytes.Equal(receivedMember.TrustBundleHash, r.MemberB.TrustBundleHash) {
					// harvester trust bundle is outdated
					response[r.MemberB.TrustDomain] = common.MemberState{
						TrustDomain:     r.MemberB.TrustDomain,
						TrustBundle:     r.MemberB.TrustBundle,
						TrustBundleHash: r.MemberB.TrustBundleHash,
					}
				}
				found = true
			}
		}
		// harvester is not aware of a new relationship
		if !found {
			if r.MemberA.ID != receivedID {
				response[r.MemberA.TrustDomain] = common.MemberState{
					TrustDomain:     r.MemberB.TrustDomain,
					TrustBundle:     r.MemberB.TrustBundle,
					TrustBundleHash: r.MemberB.TrustBundleHash,
				}
			} else {
				response[r.MemberB.TrustDomain] = common.MemberState{
					TrustDomain:     r.MemberB.TrustDomain,
					TrustBundle:     r.MemberB.TrustBundle,
					TrustBundleHash: r.MemberB.TrustBundleHash,
				}
			}
		}
	}
	return response
}

func (s *echoServer) calculateBundleState(relationships []*datastore.Relationship, receivedTD string) (common.FederationState, error) {
	response := make(common.FederationState, len(relationships))

	for _, r := range relationships {
		if r.MemberA.TrustDomain != receivedTD {
			response[r.MemberA.TrustDomain] = common.MemberState{
				TrustDomain:     r.MemberA.TrustDomain,
				TrustBundle:     r.MemberA.TrustBundle,
				TrustBundleHash: r.MemberA.TrustBundleHash,
			}
		}
		if r.MemberB.TrustDomain != receivedTD {
			response[r.MemberB.TrustDomain] = common.MemberState{
				TrustDomain:     r.MemberB.TrustDomain,
				TrustBundle:     r.MemberB.TrustBundle,
				TrustBundleHash: r.MemberB.TrustBundleHash,
			}
		}
	}

	return response, nil
}
