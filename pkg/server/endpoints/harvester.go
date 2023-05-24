package endpoints

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/common/util/encoding"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const (
	authTrustDomainKey = "trust_domain"
	authClaimsKey      = "auth_claims"
)

type HarvesterAPIHandlers struct {
	Logger       logrus.FieldLogger
	Datastore    db.Datastore
	jwtIssuer    jwt.Issuer
	jwtValidator jwt.Validator
}

// NewHarvesterAPIHandlers creates a new HarvesterAPIHandlers
func NewHarvesterAPIHandlers(l logrus.FieldLogger, ds db.Datastore, jwtIssuer jwt.Issuer, jwtValidator jwt.Validator) *HarvesterAPIHandlers {
	return &HarvesterAPIHandlers{
		Logger:       l,
		Datastore:    ds,
		jwtIssuer:    jwtIssuer,
		jwtValidator: jwtValidator,
	}
}

// GetRelationships lists all the relationships for a given trust domain name and consent status  - (GET /relationships)
// The consent status is optional, if not provided, all relationships will be returned for a given trust domain. If the
// consent status is provided, only relationships with the given consent status for the given trust domain will be returned.
// The trust domain name provided should match the authenticated trust domain.
func (h *HarvesterAPIHandlers) GetRelationships(echoCtx echo.Context, trustDomainName api.TrustDomainName, params harvester.GetRelationshipsParams) error {
	ctx := echoCtx.Request().Context()

	authTD, err := h.getAuthenticateTrustDomain(echoCtx, trustDomainName)
	if err != nil {
		return err
	}

	consentStatus := *params.ConsentStatus
	switch consentStatus {
	case "", api.Approved, api.Denied, api.Pending:
	default:
		err := fmt.Errorf("invalid consent status: %q", *params.ConsentStatus)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	// get the relationships for the trust domain
	relationships, err := h.Datastore.FindRelationshipsByTrustDomainID(ctx, authTD.ID.UUID)
	if err != nil {
		msg := "error looking up relationships"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	if consentStatus != "" {
		relationships = entity.FilterRelationships(relationships, entity.ConsentStatus(consentStatus), &authTD.ID.UUID)
	}

	relationships, err = db.PopulateTrustDomainNames(ctx, h.Datastore, relationships...)
	if err != nil {
		msg := "failed populating relationships entities"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	apiRelationships := api.MapRelationships(relationships...)

	return chttp.WriteResponse(echoCtx, http.StatusOK, apiRelationships)
}

// PatchRelationship approves/denies relationships requests - (PATCH /trust-domain/{trustDomainName}/relationships/{relationshipID})
func (h *HarvesterAPIHandlers) PatchRelationship(echoCtx echo.Context, trustDomainName api.TrustDomainName, relationshipID api.UUID) error {
	ctx := echoCtx.Request().Context()

	authTD, ok := echoCtx.Get(authTrustDomainKey).(*entity.TrustDomain)
	if !ok {
		err := errors.New("no authenticated trust domain")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusUnauthorized)
	}

	if authTD.Name.String() != trustDomainName {
		err := fmt.Errorf("trust domain name in path doesn't match authenticated trust domain name")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	// get the relationships for the trust domain
	relationship, err := h.Datastore.FindRelationshipByID(ctx, relationshipID)
	if err != nil {
		msg := "error looking up relationships"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	if relationship == nil {
		err := fmt.Errorf("relationship not found")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusNotFound)
	}

	if relationship.TrustDomainAID != authTD.ID.UUID && relationship.TrustDomainBID != authTD.ID.UUID {
		err := fmt.Errorf("relationship doesn't belong to the authenticated trust domain")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusUnauthorized)
	}

	var patchRequest harvester.PatchRelationshipRequest
	if err := chttp.ParseRequestBodyToStruct(echoCtx, &patchRequest); err != nil {
		msg := "error reading body"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}

	consentStatus := patchRequest.ConsentStatus

	switch consentStatus {
	case api.Approved, api.Denied, api.Pending:
	default:
		err := fmt.Errorf("invalid consent status: %q", consentStatus)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	// update the relationship consent status for the authenticated trust domain
	if relationship.TrustDomainAID == authTD.ID.UUID {
		relationship.TrustDomainAConsent = entity.ConsentStatus(consentStatus)
	} else {
		relationship.TrustDomainBConsent = entity.ConsentStatus(consentStatus)
	}

	updatedRel, err := h.Datastore.CreateOrUpdateRelationship(ctx, relationship)
	if err != nil {
		msg := "error updating relationship"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	r, err := db.PopulateTrustDomainNames(ctx, h.Datastore, updatedRel)
	if err != nil {
		msg := "failed populating relationships entities"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	resp := api.RelationshipFromEntity(r[0])

	if err = chttp.WriteResponse(echoCtx, http.StatusOK, resp); err != nil {
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// Onboard introduces a harvester to Galadriel Server providing its join token, and gets back a JWT token - (GET /trust-domain/onboard)
func (h *HarvesterAPIHandlers) Onboard(echoCtx echo.Context, trustDomainName api.TrustDomainName, params harvester.OnboardParams) error {
	ctx := echoCtx.Request().Context()

	if trustDomainName == "" {
		err := errors.New("trust domain name is required")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}
	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		err := fmt.Errorf("invalid trust domain name: %q", trustDomainName)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	if params.JoinToken == "" {
		err := errors.New("join token is required")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	token, err := h.Datastore.FindJoinToken(ctx, params.JoinToken)
	if err != nil {
		msg := "error looking up token"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	if token == nil {
		err := errors.New("token not found")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	if token.ExpiresAt.Before(time.Now()) {
		msg := "token expired"
		err := fmt.Errorf("%s: trust domain ID: %s", msg, token.TrustDomainID)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusUnauthorized)
	}

	if token.Used {
		msg := "token already used"
		err := fmt.Errorf("%s: trust domain name: %s", msg, trustDomainName)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}

	trustDomain, err := h.Datastore.FindTrustDomainByID(ctx, token.TrustDomainID)
	if err != nil {
		msg := "error looking up trust domain"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	if trustDomain == nil {
		msg := "trust domain not found"
		err := fmt.Errorf("%s: trust domain ID: %s", msg, token.TrustDomainID)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}

	if trustDomain.Name != tdName {
		msg := "trust domain name does not match the one associated to the token"
		err := fmt.Errorf("%s: trust domain ID: %s", msg, token.TrustDomainID)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}

	// mark token as used
	if _, err := h.Datastore.UpdateJoinToken(ctx, token.ID.UUID, true); err != nil {
		msg := "failed to update token"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	jwtParams := &jwt.JWTParams{
		Issuer:   constants.GaladrielServerName,
		Subject:  trustDomain.Name,
		Audience: []string{constants.GaladrielServerName},
		TTL:      24 * 5 * time.Hour,
	}

	jwtToken, err := h.jwtIssuer.IssueJWT(ctx, jwtParams)
	if err != nil {
		msg := "error generating JWT token"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	h.Logger.WithField(telemetry.TrustDomain, tdName.String()).Debug("Harvester onboarded successfully")

	resp := &harvester.OnboardHarvesterResponse{
		Token:           jwtToken,
		TrustDomainID:   trustDomain.ID.UUID,
		TrustDomainName: trustDomain.Name.String(),
	}

	return chttp.WriteResponse(echoCtx, http.StatusOK, resp)
}

// GetNewJWTToken renews a JWT access token - (GET /trust-domain/jwt)
func (h *HarvesterAPIHandlers) GetNewJWTToken(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	ctx := echoCtx.Request().Context()

	if trustDomainName == "" {
		err := errors.New("trust domain name is required")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}
	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		err := fmt.Errorf("invalid trust domain name: %q", trustDomainName)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	claims, ok := echoCtx.Get(authClaimsKey).(*gojwt.RegisteredClaims)
	if !ok {
		msg := "failed to parse JWT access token claims"
		err := fmt.Errorf("%s", msg)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusUnauthorized)
	}

	sub := claims.Subject
	subject, err := spiffeid.TrustDomainFromString(sub)
	if err != nil {
		msg := "failed to parse trust domain from subject"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusUnauthorized)
	}

	if subject != tdName {
		msg := "trust domain name does not match the token subject"
		err := fmt.Errorf("%s: trust domain ID: %s", msg, tdName)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}

	// params for the new JWT token
	params := jwt.JWTParams{
		Issuer: constants.GaladrielServerName,
		// the new JWT token has the same subject as the received token
		Subject:  subject,
		Audience: []string{constants.GaladrielServerName},
	}

	newToken, err := h.jwtIssuer.IssueJWT(ctx, &params)
	if err != nil {
		msg := "failed to generate new JWT token"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	jwtResp := harvester.GetJwtResponse{Token: newToken}

	h.Logger.WithField(telemetry.TrustDomain, subject).Debug("Issue new JWT token")

	return chttp.WriteResponse(echoCtx, http.StatusOK, jwtResp)
}

// BundleSync synchronizes the status of trust bundles between server and harvester - (POST /trust-domain/{trustDomainName}/bundles/sync)
func (h *HarvesterAPIHandlers) BundleSync(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	ctx := echoCtx.Request().Context()

	authTD, err := h.getAuthenticateTrustDomain(echoCtx, trustDomainName)
	if err != nil {
		return err
	}

	// Get the request body
	var req harvester.PostBundleSyncRequest
	if err := chttp.ParseRequestBodyToStruct(echoCtx, &req); err != nil {
		msg := "failed to parse request body"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}

	// Look up relationships the authenticated trust domain has with other trust domains
	relationships, err := h.Datastore.FindRelationshipsByTrustDomainID(ctx, authTD.ID.UUID)
	if err != nil {
		msg := "failed to look up relationships"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	// filer out the relationships whose consent status is not "approved" by the authenticated trust domain
	relationships = entity.FilterRelationships(relationships, entity.ConsentStatusApproved, &authTD.ID.UUID)

	resp, err := h.getBundleSyncResult(ctx, authTD, relationships, req)
	if err != nil {
		msg := "failed to generate bundle sync result"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	h.Logger.WithField(telemetry.TrustDomain, trustDomainName).Debug("Bundle sync request complete")

	return chttp.WriteResponse(echoCtx, http.StatusOK, resp)
}

// BundlePut uploads a new trust bundle to the server  - (PUT /trust-domain/{trustDomainName}/bundles)
func (h *HarvesterAPIHandlers) BundlePut(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	ctx := echoCtx.Request().Context()

	authTD, err := h.getAuthenticateTrustDomain(echoCtx, trustDomainName)
	if err != nil {
		return err
	}

	req := &harvester.BundlePutJSONRequestBody{}
	if err := chttp.ParseRequestBodyToStruct(echoCtx, req); err != nil {
		msg := "failed to read bundle from request body"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}

	if err := validateBundleRequest(req); err != nil {
		err := fmt.Errorf("invalid bundle request: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	if authTD.Name.String() != req.TrustDomain {
		err := fmt.Errorf("trust domain in request bundle %q does not match authenticated trust domain: %q", req.TrustDomain, authTD.Name.String())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusUnauthorized)
	}

	bundle, err := req.ToEntity()
	if err != nil {
		msg := "failed to parse request bundle"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusBadRequest)
	}
	// ensure that the bundle's trust domain ID matches the authenticated trust domain ID
	bundle.TrustDomainID = authTD.ID.UUID

	storedBundle, err := h.Datastore.FindBundleByTrustDomainID(ctx, authTD.ID.UUID)
	if err != nil {
		msg := "failed looking up bundle in DB"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	// the bundle already exists in the datastore, so we need to update it
	if storedBundle != nil {
		bundle.ID = storedBundle.ID
	}

	if _, err := h.Datastore.CreateOrUpdateBundle(ctx, bundle); err != nil {
		msg := "failed to store bundle in DB"
		err := fmt.Errorf("%s: %w", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	h.Logger.WithField(telemetry.TrustDomain, authTD.Name.String()).Info("Stored new bundle")

	if err = chttp.RespondWithoutBody(echoCtx, http.StatusOK); err != nil {
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (h *HarvesterAPIHandlers) getBundleSyncResult(ctx context.Context, authTD *entity.TrustDomain, relationships []*entity.Relationship, req harvester.PostBundleSyncRequest) (*harvester.PostBundleSyncResponse, error) {
	resp := &harvester.PostBundleSyncResponse{
		State:   make(map[string]api.BundleDigest, len(relationships)),
		Updates: make(harvester.BundlesUpdates),
	}

	for _, relationship := range relationships {
		relatedTrustDomainID := relationship.TrustDomainAID
		if relationship.TrustDomainAID == authTD.ID.UUID {
			relatedTrustDomainID = relationship.TrustDomainBID
		}
		bundle, err := h.Datastore.FindBundleByTrustDomainID(ctx, relatedTrustDomainID)
		if err != nil {
			return nil, err
		}

		if bundle == nil {
			// no bundle for the federated trust domain, nothing to do
			continue
		}

		td, err := h.Datastore.FindTrustDomainByID(ctx, relatedTrustDomainID)
		if err != nil {
			return nil, err
		}
		bundle.TrustDomainName = td.Name

		// Look up the bundle digest in the request
		reqDigest, ok := req.State[bundle.TrustDomainName.String()]
		decodedReqDigest, err := encoding.DecodeFromBase64(reqDigest)
		if err != nil {
			return nil, err
		}

		// The bundle digest in the request is different from the stored one, so the bundle needs to be updated
		if !ok || !bytes.Equal(bundle.Digest[:], decodedReqDigest) {
			updateItem := harvester.BundlesUpdatesItem{}
			updateItem.TrustBundle = string(bundle.Data)
			updateItem.Digest = encoding.EncodeToBase64(bundle.Digest[:])
			updateItem.Signature = encoding.EncodeToBase64(bundle.Signature)
			updateItem.SigningCertificate = encoding.EncodeToBase64(bundle.SigningCertificate)
			resp.Updates[bundle.TrustDomainName.String()] = updateItem
		}

		// Add the bundle to the current state
		resp.State[bundle.TrustDomainName.String()] = encoding.EncodeToBase64(bundle.Digest[:])
	}

	return resp, nil
}

func (h *HarvesterAPIHandlers) getAuthenticateTrustDomain(echoCtx echo.Context, trustDomainName string) (*entity.TrustDomain, error) {
	authTD, ok := echoCtx.Get(authTrustDomainKey).(*entity.TrustDomain)
	if !ok {
		err := errors.New("no authenticated trust domain")
		return nil, chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusUnauthorized)
	}

	if authTD.Name.String() != trustDomainName {
		err := fmt.Errorf("request trust domain %q does not match authenticated trust domain %q", trustDomainName, authTD.Name.String())
		return nil, chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusUnauthorized)
	}

	return authTD, nil
}

func validateBundleRequest(req *harvester.BundlePutJSONRequestBody) error {
	if req.TrustDomain == "" {
		return errors.New("bundle trust domain is required")
	}

	if req.TrustBundle == "" {
		return errors.New("trust bundle is required")
	}

	if req.Digest == "" {
		return errors.New("bundle digest is required")
	}

	decodedDigest, err := encoding.DecodeFromBase64(req.Digest)
	if err != nil {
		return fmt.Errorf("failed decoding bundle digest: %w", err)
	}
	if err := cryptoutil.ValidateBundleDigest([]byte(req.TrustBundle), decodedDigest); err != nil {
		return fmt.Errorf("failed validating bundle digest: %w", err)
	}

	return nil
}
