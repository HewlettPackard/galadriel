package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"io"
	"strings"
)

func (e *Endpoints) syncBundleHandler(ctx echo.Context) error {
	logger := ctx.Logger()
	logger.Debug("Receiving sync request")

	tokenHeader := ctx.Request().Header["Authorization"]
	token, err := e.DataStore.GetAccessToken(context.TODO(), strings.Split(tokenHeader[0], "Bearer ")[1])
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("error fetching token: %v", err))
		return err
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to read body: %v", err))
		return err
	}

	receivedHarvesterState := common.SyncBundleRequest{}
	err = json.Unmarshal(body, &receivedHarvesterState)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to unmarshal state: %v", err))
		return err
	}

	_, foundSelf := receivedHarvesterState.FederatesWith[token.Member.TrustDomain]
	if foundSelf {
		e.handleTcpError(ctx, "bad request: harvester cannot federate with itself")
		return err
	}

	// get relationships for that member
	relationships, err := e.DataStore.GetRelationships(context.TODO(), token.Member.TrustDomain)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to fetch relationships: %v", err))
		return err
	}

	response := common.SyncBundleResponse{}

	currentState, err := e.calculateBundleState(relationships, token.Member.TrustDomain)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed calculating bundle state: %v", err))
		return err
	}

	response.State = currentState
	response.Update = e.calculateBundleSync(receivedHarvesterState.FederatesWith, relationships, token.Member)

	responseBytes, err := json.Marshal(response)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to marshal response: %v", err))
		return err
	}

	_, err = ctx.Response().Write(responseBytes)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to write response: %v", err))
		return err
	}

	return nil
}

func (e *Endpoints) postBundleHandler(ctx echo.Context) error {
	logger := ctx.Logger()
	logger.Debug("Receiving post bundle request")

	tokenHeader := ctx.Request().Header["Authorization"]
	token, err := e.DataStore.GetAccessToken(context.TODO(), strings.Split(tokenHeader[0], "Bearer ")[1])
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("error fetching token: %v", err))
		return err
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to read body: %v", err))
		return err
	}

	receivedHarvesterState := common.PostBundleRequest{}
	err = json.Unmarshal(body, &receivedHarvesterState)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to unmarshal state: %v", err))
		return err
	}
	receivedHarvesterState.TrustBundleHash = calculateBundleHash(receivedHarvesterState.TrustBundle)

	// fetch harvester state from db
	currentState, err := e.DataStore.GetMember(context.TODO(), token.Member.TrustDomain)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to fetch member: %v", err))
		return err
	}

	// update internal state
	if !bytes.Equal(receivedHarvesterState.TrustBundleHash, currentState.TrustBundleHash) {
		_, err := e.DataStore.UpdateMember(context.TODO(), token.Member.TrustDomain, &common.Member{
			TrustBundle:     receivedHarvesterState.TrustBundle,
			TrustBundleHash: receivedHarvesterState.TrustBundleHash,
		})
		if err != nil {
			e.handleTcpError(ctx, fmt.Sprintf("failed to update member: %v", err))
			return err
		}
		logger.Infof("Trust domain %s has been successfully updated", receivedHarvesterState.TrustDomain)
	}

	return nil
}

// calculateBundleSync iterates over all relationships in db, and for each, it calls findOutdatedRelationship.
// Calls to findOutdatedRelationship will return if the currently iterated relationship was found or not, and also
// if it's outdated or not. This way we can know if we need to append an update request, either due to the relationship
// not being found, or by holding an outdated relationship state.
func (e *Endpoints) calculateBundleSync(
	received common.FederationState,
	current []*datastore.Relationship,
	callingMember *datastore.Member) common.FederationState {
	response := make(common.FederationState)

	for _, r := range current {
		update, found := findOutdatedRelationship(r, received, callingMember)
		if found && update != nil {
			response[update.TrustDomain] = *update
			continue
		}

		if r.MemberA.TrustDomain != callingMember.TrustDomain && r.MemberA.TrustBundle != nil {
			response[r.MemberA.TrustDomain] = common.MemberState{
				TrustDomain: r.MemberB.TrustDomain,
				TrustBundle: r.MemberB.TrustBundle,
			}
		} else if r.MemberB.TrustBundle != nil {
			response[r.MemberB.TrustDomain] = common.MemberState{
				TrustDomain: r.MemberB.TrustDomain,
				TrustBundle: r.MemberB.TrustBundle,
			}
		}
	}
	return response
}

// findOutdatedRelationship tries to find a federated entry found in r, in the received FederationState.
// If we find a match, we validate it's state and return the updated version.
// A (nil, false) response means no update is needed, but creation is needed since we did not find it.
// A (non-nil, true) response means we found a match, but its outdated, so update it with the returned MemberState.
func findOutdatedRelationship(
	r *datastore.Relationship,
	received common.FederationState,
	callingMember *datastore.Member) (*common.MemberState, bool) {
	if r.MemberA.TrustDomain != callingMember.TrustDomain {
		member, ok := received[r.MemberA.TrustDomain]
		if !ok {
			return nil, false
		}
		if !bytes.Equal(r.MemberA.TrustBundleHash, member.TrustBundleHash) {
			return &common.MemberState{
				TrustDomain: r.MemberA.TrustDomain,
				TrustBundle: r.MemberA.TrustBundle,
			}, true
		}
		return nil, true
	}

	member, ok := received[r.MemberB.TrustDomain]
	if !ok {
		return nil, false
	}
	if !bytes.Equal(r.MemberB.TrustBundleHash, member.TrustBundleHash) {
		return &common.MemberState{
			TrustDomain: r.MemberB.TrustDomain,
			TrustBundle: r.MemberB.TrustBundle,
		}, true
	}

	return nil, true
}

func (e *Endpoints) calculateBundleState(
	relationships []*datastore.Relationship,
	receivedTD string) (common.FederationState, error) {
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

func (e *Endpoints) handleTcpError(ctx echo.Context, errMsg string) {
	e.Log.Errorf(errMsg)
	_, err := ctx.Response().Write([]byte(errMsg))
	if err != nil {
		e.Log.Errorf("Failed to write error response: %v", err)
	}
}

func calculateBundleHash(bundle []byte) []byte {
	// TODO
	return bundle
}
