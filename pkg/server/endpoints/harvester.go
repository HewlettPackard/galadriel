package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

func (e *Endpoints) postBundleHandler(ctx echo.Context) error {
	e.Logger.Debug("Receiving post bundle request")

	t, ok := ctx.Get("token").(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing token")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	token, err := e.Datastore.FindJoinToken(ctx.Request().Context(), t.Token)
	if err != nil {
		err := errors.New("error looking up token")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	authenticatedTD, err := e.Datastore.FindTrustDomainByID(ctx.Request().Context(), token.TrustDomainID)
	if err != nil {
		err := errors.New("error looking up trust domain")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to read body: %v", err))
		return err
	}

	harvesterReq := common.PostBundleRequest{}
	err = json.Unmarshal(body, &harvesterReq)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to unmarshal state: %v", err))
		return err
	}

	if harvesterReq.TrustDomainName != authenticatedTD.Name {
		err := fmt.Errorf("authenticated trust domain {%s} does not match trust domain in request: {%s}", harvesterReq.TrustDomainID, token.TrustDomainID)
		e.handleTcpError(ctx, err.Error())
		return err
	}

	bundle, err := spiffebundle.Parse(authenticatedTD.Name, harvesterReq.Bundle.Data)
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

	if !bytes.Equal(harvesterReq.Digest, digest) {
		err := errors.New("calculated digest does not match received digest")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	currentStoredBundle, err := e.Datastore.FindBundleByTrustDomainID(ctx.Request().Context(), authenticatedTD.ID.UUID)
	if err != nil {
		e.handleTcpError(ctx, err.Error())
		return err
	}

	if harvesterReq.Bundle != nil && currentStoredBundle != nil && !bytes.Equal(harvesterReq.Bundle.Digest, currentStoredBundle.Digest) {
		_, err := e.Datastore.CreateOrUpdateBundle(ctx.Request().Context(), &entity.Bundle{
			Data: harvesterReq.Bundle.Data,
		})
		if err != nil {
			e.handleTcpError(ctx, fmt.Sprintf("failed to update trustDomain: %v", err))
			return err
		}

		e.Logger.Infof("Trust domain %s has been successfully updated", authenticatedTD.Name)
	} else if currentStoredBundle == nil {
		_, err := e.Datastore.CreateOrUpdateBundle(ctx.Request().Context(), &entity.Bundle{
			Data:          harvesterReq.Bundle.Data,
			Digest:        harvesterReq.Bundle.Digest,
			TrustDomainID: authenticatedTD.ID.UUID,
		})
		if err != nil {
			e.handleTcpError(ctx, fmt.Sprintf("failed to update trustDomain: %v", err))
			return err
		}

		e.Logger.Debugf("Trust domain %s has been successfully updated", receivedHarvesterState.TrustDomain)
	}

	return nil
}

func (e *Endpoints) syncFederatedBundleHandler(ctx echo.Context) error {
	e.Logger.Debug("Receiving sync request")

	t, ok := ctx.Get("token").(*entity.JoinToken)
	if !ok {
		err := errors.New("error parsing join token")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	token, err := e.Datastore.FindJoinToken(ctx.Request().Context(), t.Token)
	if !ok {
		err := errors.New("error looking up token")
		e.handleTcpError(ctx, err.Error())
		return err
	}

	harvesterTrustDomain, err := e.Datastore.FindTrustDomainByID(ctx.Request().Context(), token.TrustDomainID)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to read body: %v", err))
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

	harvesterBundleDigests := receivedHarvesterState.State

	_, foundSelf := receivedHarvesterState.State[harvesterTrustDomain.Name]
	if foundSelf {
		e.handleTcpError(ctx, "bad request: harvester cannot federate with itself")
		return err
	}

	relationships, err := e.Datastore.FindRelationshipsByTrustDomainID(ctx.Request().Context(), harvesterTrustDomain.ID.UUID)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to fetch relationships: %v", err))
		return err
	}

	federatedTDs := getFederatedTrustDomains(relationships, harvesterTrustDomain.ID.UUID)

	if len(federatedTDs) == 0 {
		e.Logger.Info("No federated trust domains yet")
		return nil
	}

	federatedBundles, federatedBundlesDigests, err := e.getCurrentFederatedBundles(ctx.Request().Context(), federatedTDs)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to fetch bundles from DB: %v", err))
		return err
	}

	if len(federatedBundles) == 0 {
		e.Logger.Info("No federated bundles yet")
		return nil
	}

	bundlesUpdates, err := e.getFederatedBundlesUpdates(ctx.Request().Context(), harvesterBundleDigests, federatedBundles)
	if err != nil {
		e.handleTcpError(ctx, fmt.Sprintf("failed to fetch bundles from DB: %v", err))
		return err
	}

	response := common.SyncBundleResponse{
		Updates: bundlesUpdates,
		State:   federatedBundlesDigests,
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

func getFederatedTrustDomains(relationships []*entity.Relationship, tdID uuid.UUID) []uuid.UUID {
	var federatedTrustDomains []uuid.UUID

	for _, r := range relationships {
		ma := r.TrustDomainAID
		mb := r.TrustDomainBID

		if tdID == ma {
			federatedTrustDomains = append(federatedTrustDomains, mb)
		} else {
			federatedTrustDomains = append(federatedTrustDomains, ma)
		}
	}
	return federatedTrustDomains
}

func (e *Endpoints) getFederatedBundlesUpdates(ctx context.Context, harvesterBundlesDigests common.BundlesDigests, federatedBundles []*entity.Bundle) (common.BundleUpdates, error) {
	response := make(common.BundleUpdates)

	for _, b := range federatedBundles {
		td, err := e.Datastore.FindTrustDomainByID(ctx, b.TrustDomainID)
		if err != nil {
			return nil, err
		}

		serverDigest := b.Digest
		harvesterDigest := harvesterBundlesDigests[td.Name]

		// If the bundle digest received from a federated trust domain of the calling harvester is not the same as the
		// digest the server has, the harvester needs to be updated of the new bundle. This also covers the case of
		// the harvester not being aware of any bundles. The update represents a newly federated trustDomain's bundle.
		if !bytes.Equal(harvesterDigest, serverDigest) {
			response[td.Name] = b
		}
	}

	return response, nil
}

func (e *Endpoints) getCurrentFederatedBundles(ctx context.Context, federatedTDs []uuid.UUID) ([]*entity.Bundle, common.BundlesDigests, error) {
	var bundles []*entity.Bundle
	bundlesDigests := map[spiffeid.TrustDomain][]byte{}

	for _, id := range federatedTDs {
		b, err := e.Datastore.FindBundleByTrustDomainID(ctx, id)
		if err != nil {
			return nil, nil, err
		}
		td, err := e.Datastore.FindTrustDomainByID(ctx, id)
		if err != nil {
			return nil, nil, err
		}

		if b != nil {
			bundles = append(bundles, b)
			bundlesDigests[td.Name] = b.Digest
		}
	}

	return bundles, bundlesDigests, nil
}

func (e *Endpoints) handleTcpError(ctx echo.Context, errMsg string) {
	e.Logger.Errorf(errMsg)
	_, err := ctx.Response().Write([]byte(errMsg))
	if err != nil {
		e.Logger.Errorf("Failed to write error response: %v", err)
	}
}
