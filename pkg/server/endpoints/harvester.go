package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/labstack/echo/v4"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"io"
)

func (e *Endpoints) syncFederatedBundleHandler(ctx echo.Context) error {
	e.Log.Debug("Receiving sync request")

	token, ok := ctx.Get("token").(*common.AccessToken)
	if !ok {
		err := errors.New("error asserting user token")
		e.handleTcpError(ctx, err.Error())
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

	_, foundSelf := receivedHarvesterState.State[token.TrustDomain]
	if foundSelf {
		e.handleTcpError(ctx, "bad request: harvester cannot federate with itself")
		return err
	}

	relationships, err := e.DataStore.GetRelationships(context.TODO(), token.TrustDomain.String())
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to fetch relationships: %v", err))
		return err
	}
	response := common.SyncBundleResponse{}

	currentState, err := e.calculateBundleState(relationships, token.TrustDomain)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed calculating bundle state: %v", err))
		return err
	}

	response.State = currentState
	response.Updates = e.calculateBundleSync(receivedHarvesterState.State, relationships, token.TrustDomain)

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
	e.Log.Debug("Receiving post bundle request")

	token, ok := ctx.Get("token").(*common.AccessToken)
	if !ok {
		err := errors.New("error asserting user token")
		e.handleTcpError(ctx, err.Error())
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

	if receivedHarvesterState.TrustDomain.Compare(token.TrustDomain) != 0 {
		err := errors.New("authenticated trust domain does not match received trust domain")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	if !bytes.Equal(receivedHarvesterState.BundleDigest, util.GetDigest(receivedHarvesterState.Bundle)) {
		err := errors.New("calculated digest does not match received digest")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	currentState, err := e.DataStore.GetMember(context.TODO(), token.TrustDomain.String())
	if err != nil {
		e.handleTcpError(ctx, err.Error())
		return err
	}

	if !bytes.Equal(receivedHarvesterState.BundleDigest, currentState.BundleDigest) {
		_, err := e.DataStore.UpdateMember(context.TODO(), token.TrustDomain.String(), &common.Member{
			TrustBundle: receivedHarvesterState.TrustBundle,
		})
		if err != nil {
			e.handleTcpError(ctx, fmt.Sprintf("failed to update member: %v", err))
			return err
		}
		e.Log.Debugf("Trust domain %s has been successfully updated", receivedHarvesterState.TrustDomain)
	}

	return nil
}

// calculateBundleSync iterates over all relationships in db, and for each, it calls findOutdatedRelationship. These
// calls return if the currently iterated relationship was found or not, and its updated version, in case it was found
// outdated. This way we can know if we need to append an update request, either due to the relationship not being
// found, or by finding an outdated relationship state.
func (e *Endpoints) calculateBundleSync(
	received common.BundlesDigests,
	current []*common.Relationship,
	callerTD spiffeid.TrustDomain) common.BundleUpdates {
	response := make(common.BundleUpdates)

	for _, r := range current {
		update, found := e.findOutdatedRelationship(r, received, callerTD)

		if !found {
			// trust bundle could be nil in case we didn't receive one yet.
			if r.MemberA.TrustDomain.Compare(callerTD) != 0 && r.MemberA.Bundle != nil {
				response[r.MemberA.TrustDomain] = r.MemberA.TrustBundle
			} else if r.MemberB.Bundle != nil {
				response[r.MemberB.TrustDomain] = r.MemberB.TrustBundle
			}

			continue
		}

		if update != nil {
			response[update.TrustDomain] = *update
		}
	}

	return response
}

// findOutdatedRelationship looks for a federated entry found in r, in the received State.
// If we find a match, we validate its state and return the updated version.
// A (nil, false) response means no update is needed, but creation is needed since we did not find it.
// A (non-nil, true) response means we found a match, but its outdated, so update it with the returned TrustBundle.
func (e *Endpoints) findOutdatedRelationship(
	r *common.Relationship,
	received common.BundlesDigests,
	callerTD spiffeid.TrustDomain) (*common.TrustBundle, bool) {
	if r.MemberA.TrustDomain.Compare(callerTD) != 0 {
		receivedDigest, ok := received[r.MemberA.TrustDomain]
		if !ok {
			return nil, false
		}

		if !bytes.Equal(r.MemberA.BundleDigest, receivedDigest) {
			return &r.MemberA.TrustBundle, true
		}

		return nil, true
	}

	if r.MemberB.TrustDomain.String() != callerTD.String() {
		receivedDigest, ok := received[r.MemberB.TrustDomain]
		if !ok {
			return nil, false
		}
		if !bytes.Equal(r.MemberB.BundleDigest, receivedDigest) {
			return &r.MemberB.TrustBundle, true
		}

		return nil, true
	}

	// TODO: sanity check for this state
	return nil, false
}

func (e *Endpoints) calculateBundleState(
	relationships []*common.Relationship,
	callerTD spiffeid.TrustDomain) (common.BundlesDigests, error) {
	federationState := make(common.BundlesDigests, len(relationships))

	for _, r := range relationships {
		if r.MemberA.TrustDomain.Compare(callerTD) != 0 {
			federationState[r.MemberA.TrustDomain] = r.MemberA.BundleDigest
		}
		if r.MemberB.TrustDomain.Compare(callerTD) != 0 {
			federationState[r.MemberB.TrustDomain] = r.MemberB.BundleDigest
		}
	}

	return federationState, nil
}

func (e *Endpoints) handleTcpError(ctx echo.Context, errMsg string) {
	e.Log.Errorf(errMsg)
	_, err := ctx.Response().Write([]byte(errMsg))
	if err != nil {
		e.Log.Errorf("Failed to write error response: %v", err)
	}
}
