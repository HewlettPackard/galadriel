package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/labstack/echo/v4"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (e *Endpoints) postBundleHandler(ctx echo.Context) error {
	e.Logger.Debug("Receiving post bundle request")

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
		err := fmt.Errorf("authenticated trust domain {%s} does not match received trust domain {%s}", receivedHarvesterState.TrustDomain.String(), token.TrustDomain.String())

		e.handleTcpError(ctx, err.Error())
		return err
	}

	bundle, err := spiffebundle.Parse(receivedHarvesterState.TrustDomain, receivedHarvesterState.Bundle)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to parse bundle: %v", err))
		return err
	}

	x509b, err := bundle.X509Bundle().Marshal()
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to marshal bundle: %v", err))
		return err
	}

	digest := util.GetDigest(x509b)

	if !bytes.Equal(receivedHarvesterState.BundleDigest, digest) {
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

		e.Logger.Debugf("Trust domain %s has been successfully updated", receivedHarvesterState.TrustDomain)
	}

	return nil
}

func (e *Endpoints) syncFederatedBundleHandler(ctx echo.Context) error {
	e.Logger.Debug("Receiving sync request")

	token, ok := ctx.Get("token").(*common.AccessToken)
	if !ok {
		err := errors.New("error asserting user token")
		e.handleTcpError(ctx, err.Error())
		return err
	}
	harvesterTrustDomain := token.TrustDomain

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

	harvesterBundleDigests := receivedHarvesterState.State

	_, foundSelf := receivedHarvesterState.State[token.TrustDomain]
	if foundSelf {
		e.handleTcpError(ctx, "bad request: harvester cannot federate with itself")
		return err
	}

	relationships, err := e.DataStore.GetRelationships(context.TODO(), harvesterTrustDomain.String())
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to fetch relationships: %v", err))
		return err
	}

	federatedMembers := getFederatedMembers(relationships, harvesterTrustDomain)
	lastBundlesDigests := getCurrentFederatedBundleDigests(federatedMembers)
	bundlesUpdates := getFederatedBundlesUpdates(harvesterBundleDigests, federatedMembers)

	response := common.SyncBundleResponse{
		Updates: bundlesUpdates,
		State:   lastBundlesDigests,
	}

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

func getFederatedMembers(relationships []*common.Relationship, td spiffeid.TrustDomain) []*common.Member {
	var federatedMembers []*common.Member

	for _, r := range relationships {
		ma := r.MemberA
		mb := r.MemberB

		if td.Compare(ma.TrustDomain) != 0 {
			federatedMembers = append(federatedMembers, ma)
		} else {
			federatedMembers = append(federatedMembers, mb)
		}
	}
	return federatedMembers
}

func getFederatedBundlesUpdates(harvesterBundlesDigests common.BundlesDigests, federatedMembers []*common.Member) common.BundleUpdates {
	response := make(common.BundleUpdates)

	for _, m := range federatedMembers {
		td := m.TrustDomain
		serverDigest := m.BundleDigest
		harvesterDigest := harvesterBundlesDigests[td]

		// If the bundle digest received from a federated trust domain of the calling harvester is not the same as the
		// digest the server has, the harvester needs to be updated of the new bundle. This also covers the case of
		// the harvester not being aware of any bundles. The update represents a newly federated member's bundle.
		if !bytes.Equal(harvesterDigest, serverDigest) {
			tb := common.TrustBundle{
				TrustDomain:  td,
				Bundle:       m.Bundle,
				BundleDigest: m.BundleDigest,
			}
			response[td] = tb
		}
	}

	return response
}

func getCurrentFederatedBundleDigests(federatedMembers []*common.Member) common.BundlesDigests {
	bundlesDigests := make(common.BundlesDigests, len(federatedMembers))
	for _, m := range federatedMembers {
		bundlesDigests[m.TrustDomain] = m.BundleDigest
	}
	return bundlesDigests
}

func (e *Endpoints) handleTcpError(ctx echo.Context, errMsg string) {
	e.Logger.Errorf(errMsg)
	_, err := ctx.Response().Write([]byte(errMsg))
	if err != nil {
		e.Logger.Errorf("Failed to write error response: %v", err)
	}
}
