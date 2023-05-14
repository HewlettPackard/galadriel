package endpoints

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/constants"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/jwt"
	"github.com/HewlettPackard/galadriel/pkg/server/api/harvester"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	gojwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

const (
	authTrustDomainKey = "trust_domain"
	authClaimsKey      = "auth_claims"
	defaultJWTTTL      = 24 * 5 * time.Hour
)

type HarvesterAPIHandlers struct {
	Logger       logrus.FieldLogger
	Datastore    datastore.Datastore
	jwtIssuer    jwt.Issuer
	jwtValidator jwt.Validator
}

// NewHarvesterAPIHandlers create a new HarvesterAPIHandlers
func NewHarvesterAPIHandlers(l logrus.FieldLogger, ds datastore.Datastore, jwtIssuer jwt.Issuer, jwtValidator jwt.Validator) *HarvesterAPIHandlers {
	return &HarvesterAPIHandlers{
		Logger:       l,
		Datastore:    ds,
		jwtIssuer:    jwtIssuer,
		jwtValidator: jwtValidator,
	}
}

// GetRelationships list all the relationships for a given trust domain name and consent status  - (GET /relationships)
// The consent status is optional, if not provided, all relationships will be returned for a given trust domain. If the
// consent status is provided, only relationships with the given consent status for the given trust domain will be returned.
// The trust domain name provided should match the authenticated trust domain.
func (h *HarvesterAPIHandlers) GetRelationships(echoCtx echo.Context, params harvester.GetRelationshipsParams) error {
	ctx := echoCtx.Request().Context()

	authTD, ok := echoCtx.Get(authTrustDomainKey).(*entity.TrustDomain)
	if !ok {
		err := errors.New("no authenticated trust domain")
		return h.handleErrorAndLog(err, err.Error(), http.StatusUnauthorized)
	}

	if authTD.Name.String() != params.TrustDomainName {
		err := fmt.Errorf("request trust domain %q does not match authenticated trust domain %q", params.TrustDomainName, authTD.Name.String())
		return h.handleErrorAndLog(err, err.Error(), http.StatusUnauthorized)
	}

	consentStatus := *params.ConsentStatus

	switch consentStatus {
	case "", api.Accepted, api.Denied, api.Pending:
	default:
		err := fmt.Errorf("invalid consent status: %q", *params.ConsentStatus)
		return h.handleErrorAndLog(err, err.Error(), http.StatusBadRequest)
	}

	// get the relationships for the trust domain
	relationships, err := h.Datastore.FindRelationshipsByTrustDomainID(ctx, authTD.ID.UUID)
	if err != nil {
		msg := "error looking up relationships"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	if consentStatus != "" {
		relationships = filterRelationshipsByConsentStatus(authTD.ID.UUID, relationships, consentStatus)
	}

	apiRelationships := make([]*api.Relationship, 0, len(relationships))
	for _, r := range relationships {
		apiRelationships = append(apiRelationships, api.RelationshipFromEntity(r))
	}

	return chttp.WriteResponse(echoCtx, apiRelationships)
}

// PatchRelationshipsRelationshipID accept/denied relationships requests - (PATCH /relationships/{relationshipID})
func (h *HarvesterAPIHandlers) PatchRelationshipsRelationshipID(ctx echo.Context, relationshipID api.UUID) error {
	return nil
}

// Onboard introduces a harvester to Galadriel Server providing its join token, and gets back a JWT token - (GET /trust-domain/onboard)
func (h *HarvesterAPIHandlers) Onboard(echoCtx echo.Context, params harvester.OnboardParams) error {
	ctx := echoCtx.Request().Context()

	if params.JoinToken == "" {
		err := errors.New("join token is required")
		return h.handleErrorAndLog(err, err.Error(), http.StatusBadRequest)
	}

	token, err := h.Datastore.FindJoinToken(ctx, params.JoinToken)
	if err != nil {
		msg := "error looking up token"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusBadRequest)
	}

	if token == nil {
		err := errors.New("token not found")
		return h.handleErrorAndLog(err, err.Error(), http.StatusBadRequest)
	}

	if token.ExpiresAt.Before(time.Now()) {
		msg := "token expired"
		err := fmt.Errorf("%s: trust domain ID: %s", msg, token.TrustDomainID)
		return h.handleErrorAndLog(err, msg, http.StatusBadRequest)
	}

	if token.Used {
		msg := "token already used"
		err := fmt.Errorf("%s: trust domain ID: %s", msg, token.TrustDomainID)
		return h.handleErrorAndLog(err, msg, http.StatusBadRequest)
	}

	trustDomain, err := h.Datastore.FindTrustDomainByID(ctx, token.TrustDomainID)
	if err != nil {
		msg := "error looking up trust domain"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusBadRequest)
	}

	// mark token as used
	_, err = h.Datastore.UpdateJoinToken(ctx, token.ID.UUID, true)
	if err != nil {
		msg := "failed to update token"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	h.Logger.Infof("Received onboard request from trust domain: %s", trustDomain.Name)

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
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	return chttp.WriteResponse(echoCtx, jwtToken)
}

// GetNewJWTToken renews a JWT access token - (GET /trust-domain/jwt)
func (h *HarvesterAPIHandlers) GetNewJWTToken(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	claims, ok := echoCtx.Get(authClaimsKey).(*gojwt.RegisteredClaims)
	if !ok {
		msg := "failed to parse JWT access token claims"
		err := fmt.Errorf("%s", msg)
		return h.handleErrorAndLog(err, msg, http.StatusUnauthorized)
	}

	sub := claims.Subject
	subject, err := spiffeid.TrustDomainFromString(sub)
	if err != nil {
		msg := "failed to parse trust domain from subject"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusUnauthorized)
	}

	h.Logger.Infof("New JWT token requested for trust domain %s", subject)

	// params for the new JWT token
	params := jwt.JWTParams{
		Issuer: constants.GaladrielServerName,
		// the new JWT token has the same subject as the received token
		Subject:  subject,
		Audience: []string{constants.GaladrielServerName},
		TTL:      defaultJWTTTL,
	}

	newToken, err := h.jwtIssuer.IssueJWT(ctx, &params)
	if err != nil {
		msg := "failed to generate new JWT token"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	return echoCtx.JSON(http.StatusOK, newToken)
}

// BundleSync synchronize the status of trust bundles between server and harvester - (POST /trust-domain/{trustDomainName}/bundles/sync)
func (h *HarvesterAPIHandlers) BundleSync(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	h.Logger.Debugf("Received bundle sync request from trust domain: %s", trustDomainName)
	ctx := echoCtx.Request().Context()

	// Get the authenticated trust domain from the context
	authTD, ok := echoCtx.Get(authTrustDomainKey).(*entity.TrustDomain)
	if !ok {
		err := errors.New("no authenticated trust domain")
		return h.handleErrorAndLog(err, err.Error(), http.StatusUnauthorized)
	}

	if authTD.Name.String() != trustDomainName {
		err := fmt.Errorf("request trust domain %q does not match authenticated trust domain %q", trustDomainName, authTD.Name.String())
		return h.handleErrorAndLog(err, err.Error(), http.StatusUnauthorized)
	}

	// Get the request body
	var req harvester.BundleSyncBody
	if err := echoCtx.Bind(&req); err != nil {
		msg := "failed to parse request body"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusBadRequest)
	}

	// Look up relationships the authenticated trust domain has with other trust domains
	relationships, err := h.Datastore.FindRelationshipsByTrustDomainID(ctx, authTD.ID.UUID)
	if err != nil {
		msg := "failed to look up relationships"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	// filer out the relationships whose consent status is not "accepted" by the authenticated trust domain
	relationships = filterRelationshipsByConsentStatus(authTD.ID.UUID, relationships, api.Accepted)

	resp, err := h.getBundleSyncResult(ctx, authTD, relationships, req)
	if err != nil {
		msg := "failed to generate bundle sync result"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	return echoCtx.JSON(http.StatusOK, resp)
}

// BundlePut uploads a new trust bundle to the server  - (PUT /trust-domain/{trustDomainName}/bundles)
func (h *HarvesterAPIHandlers) BundlePut(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	h.Logger.Infof("Received post bundle request from trust domain: %s", trustDomainName)
	ctx := echoCtx.Request().Context()

	// get the authenticated trust domain from the context
	authTD, ok := echoCtx.Get(authTrustDomainKey).(*entity.TrustDomain)
	if !ok {
		err := errors.New("no authenticated trust domain")
		return h.handleErrorAndLog(err, err.Error(), http.StatusUnauthorized)
	}

	if authTD.Name.String() != trustDomainName {
		err := fmt.Errorf("request trust domain %q does not match authenticated trust domain %q", trustDomainName, authTD.Name.String())
		return h.handleErrorAndLog(err, err.Error(), http.StatusUnauthorized)
	}

	req := &harvester.BundlePutJSONRequestBody{}
	err := chttp.FromBody(echoCtx, req)
	if err != nil {
		msg := "failed to read bundle from request body"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusBadRequest)
	}

	err = validateBundleRequest(req)
	if err != nil {
		err := fmt.Errorf("invalid bundle request: %v", err)
		return h.handleErrorAndLog(err, err.Error(), http.StatusBadRequest)
	}

	if authTD.Name.String() != req.TrustDomain {
		err := fmt.Errorf("trust domain in request bundle %q does not match authenticated trust domain: %q", req.TrustDomain, authTD.Name.String())
		return h.handleErrorAndLog(err, err.Error(), http.StatusUnauthorized)
	}

	bundle, err := req.ToEntity()
	if err != nil {
		msg := "failed to parse request bundle"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusBadRequest)
	}
	// ensure that the bundle's trust domain ID matches the authenticated trust domain ID
	bundle.TrustDomainID = authTD.ID.UUID

	storedBundle, err := h.Datastore.FindBundleByTrustDomainID(ctx, authTD.ID.UUID)
	if err != nil {
		msg := "failed looking up bundle in DB"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	// the bundle already exists in the datastore, so we need to update it
	if storedBundle != nil {
		bundle.ID = storedBundle.ID
	}

	_, err = h.Datastore.CreateOrUpdateBundle(ctx, bundle)
	if err != nil {
		msg := "failed to store bundle in DB"
		err := fmt.Errorf("%s: %w", msg, err)
		return h.handleErrorAndLog(err, msg, http.StatusInternalServerError)
	}

	if err = chttp.BodylessResponse(echoCtx); err != nil {
		return h.handleErrorAndLog(err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (h *HarvesterAPIHandlers) getBundleSyncResult(ctx context.Context, authTD *entity.TrustDomain, relationships []*entity.Relationship, req harvester.BundleSyncBody) (*harvester.BundleSyncResult, error) {
	resp := &harvester.BundleSyncResult{
		State:   make(map[string]api.BundleDigest, len(relationships)),
		Updates: make(harvester.TrustBundleSync),
	}

	for _, relationship := range relationships {
		otherID := relationship.TrustDomainAID
		if relationship.TrustDomainAID == authTD.ID.UUID {
			otherID = relationship.TrustDomainBID
		}
		bundle, err := h.Datastore.FindBundleByTrustDomainID(ctx, otherID)
		if err != nil {
			return nil, err
		}

		// Calculate the sha256 digest of the stored bundle
		digest := sha256.Sum256(bundle.Data)
		strDigest := encodeToBase64(digest[:])

		// Look up the bundle digest in the request
		reqDigest, _ := req.State[bundle.TrustDomainName.String()]

		if strDigest != reqDigest {
			// The bundle digest in the request is different from the stored one, so the bundle needs to be updated
			updateItem := harvester.TrustBundleSyncItem{}
			updateItem.TrustBundle = encodeToBase64(bundle.Data)
			updateItem.Signature = encodeToBase64(bundle.Signature)
			resp.Updates[bundle.TrustDomainName.String()] = updateItem
		}

		// Add the bundle to the current state
		resp.State[bundle.TrustDomainName.String()] = encodeToBase64(digest[:])
	}

	return resp, nil
}

// handleErrorAndLog logs the error and returns an HTTP error with a unique error ID which can be used to trace the error
func (h *HarvesterAPIHandlers) handleErrorAndLog(logErr error, message string, code int) error {
	errID := uuid.NewString()
	logMsg := fmt.Sprintf("%v (error ID: %s)", logErr, errID)
	errMsg := fmt.Sprintf("%s (error ID: %s)", message, errID)
	h.Logger.Errorf(logMsg)
	return echo.NewHTTPError(code, errMsg)
}

func filterRelationshipsByConsentStatus(trustDomainID uuid.UUID, relationships []*entity.Relationship, status api.ConsentStatus) []*entity.Relationship {
	filtered := make([]*entity.Relationship, 0)
	for _, r := range relationships {
		if r.TrustDomainAID == trustDomainID && api.ConsentStatus(r.TrustDomainAConsent) == status ||
			r.TrustDomainBID == trustDomainID && api.ConsentStatus(r.TrustDomainBConsent) == status {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func validateBundleRequest(req *harvester.BundlePutJSONRequestBody) error {
	if req.TrustDomain == "" {
		return errors.New("bundle trust domain is required")
	}

	if req.TrustBundle == "" {
		return errors.New("trust bundle is required")
	}

	if req.Signature == "" {
		return errors.New("bundle signature is required")
	}

	return nil
}

func encodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
