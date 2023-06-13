package endpoints

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/HewlettPackard/galadriel/pkg/server/db/criteria"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

// DefaultTokenTTL is the default TTL for tokens in seconds.
const DefaultTokenTTL = 600

type AdminAPIHandlers struct {
	Logger    logrus.FieldLogger
	Datastore db.Datastore
}

// NewAdminAPIHandlers creates a new NewAdminAPIHandlers
func NewAdminAPIHandlers(l logrus.FieldLogger, ds db.Datastore) *AdminAPIHandlers {
	return &AdminAPIHandlers{
		Logger:    l,
		Datastore: ds,
	}
}

// ListRelationships lists all relationships filtered by the request params - (GET /relationships)
func (h *AdminAPIHandlers) ListRelationships(echoCtx echo.Context, params admin.ListRelationshipsParams) error {
	ctx := echoCtx.Request().Context()

	listCriteria := &criteria.ListRelationshipsCriteria{
		OrderByCreatedAt: criteria.OrderDescending,
	}

	if params.TrustDomainName != nil {
		td, err := h.findTrustDomainByName(ctx, *params.TrustDomainName)
		if err != nil {
			err = fmt.Errorf("failed looking up trust domain name: %w", err)
			return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
		}

		listCriteria.FilterByTrustDomainID = uuid.NullUUID{Valid: true, UUID: td.ID.UUID}
	}

	var consentStatus *entity.ConsentStatus
	if params.Status != nil {
		switch *params.Status {
		case "":
		case api.Approved, api.Denied, api.Pending:
			consentStatus = (*entity.ConsentStatus)(params.Status)
		default:
			err := fmt.Errorf("status filter %q is not supported, approved values [approved, denied, pending]", *params.Status)
			return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
		}

		listCriteria.FilterByConsentStatus = consentStatus
	}

	relationships, err := h.Datastore.ListRelationships(ctx, listCriteria)
	if err != nil {
		err = fmt.Errorf("failed listing relationships: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	relationships, err = db.PopulateTrustDomainNames(ctx, h.Datastore, relationships...)
	if err != nil {
		err = fmt.Errorf("failed populating relationships entities: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	cRelationships := api.MapRelationships(relationships...)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, cRelationships)
	if err != nil {
		err = fmt.Errorf("relationships entities - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// PutRelationship creates a new relationship request between two trust domains - (PUT /relationships)
func (h *AdminAPIHandlers) PutRelationship(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	reqBody := &admin.PutRelationshipJSONRequestBody{}
	err := chttp.ParseRequestBodyToStruct(echoCtx, reqBody)
	if err != nil {
		msg := "failed to read relationship put body"
		err = fmt.Errorf("%s: %v", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}
	eRelationship, err := reqBody.ToEntity()
	if err != nil {
		return err
	}

	dbTd1, err := h.lookupTrustDomain(ctx, eRelationship.TrustDomainAName.String())
	if err != nil {
		return err
	}

	dbTd2, err := h.lookupTrustDomain(ctx, eRelationship.TrustDomainBName.String())
	if err != nil {
		return err
	}

	eRelationship.TrustDomainAID = dbTd1.ID.UUID
	eRelationship.TrustDomainBID = dbTd2.ID.UUID

	rel, err := h.Datastore.CreateOrUpdateRelationship(ctx, eRelationship)
	if err != nil {
		msg := "failed creating relationship"
		err = fmt.Errorf("%s: %v", msg, err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	h.Logger.Printf("Created relationship between trust domains %s and %s", dbTd1.Name.String(), dbTd2.Name.String())

	response := api.RelationshipFromEntity(rel)
	err = chttp.WriteResponse(echoCtx, http.StatusCreated, response)
	if err != nil {
		err = fmt.Errorf("relationships - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// GetRelationshipByID retrieves a specific relationship based on its id - (GET /relationships/{relationshipID})
func (h *AdminAPIHandlers) GetRelationshipByID(echoCtx echo.Context, relationshipID api.UUID) error {
	ctx := echoCtx.Request().Context()

	r, err := h.Datastore.FindRelationshipByID(ctx, relationshipID)
	if err != nil {
		err = fmt.Errorf("failed getting relationships: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	if r == nil {
		err = errors.New("relationship not found")
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusNotFound)
	}

	response := api.RelationshipFromEntity(r)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("relationship entity - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// PutTrustDomain creates a new trust domain - (PUT /trust-domain)
func (h *AdminAPIHandlers) PutTrustDomain(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	reqBody := &admin.PutTrustDomainJSONRequestBody{}
	err := chttp.ParseRequestBodyToStruct(echoCtx, reqBody)
	if err != nil {
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	dbTD, err := reqBody.ToEntity()
	if err != nil {
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, dbTD.Name)
	if err != nil {
		err = fmt.Errorf("failed looking up trust domain: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	if td != nil {
		err = fmt.Errorf("trust domain already exists: %q", dbTD.Name)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	m, err := h.Datastore.CreateOrUpdateTrustDomain(ctx, dbTD)
	if err != nil {
		err = fmt.Errorf("failed creating trustDomain: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	h.Logger.Printf("Created trustDomain: %s", dbTD.Name.String())

	response := api.TrustDomainFromEntity(m)
	err = chttp.WriteResponse(echoCtx, http.StatusCreated, response)
	if err != nil {
		err = fmt.Errorf("trustDomain entity - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// GetTrustDomainByName retrieves a specific trust domain by its name - (GET /trust-domain/{trustDomainName})
func (h *AdminAPIHandlers) GetTrustDomainByName(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	ctx := echoCtx.Request().Context()

	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		err = fmt.Errorf("failed parsing trust domain name: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		err = fmt.Errorf("failed getting trust domain: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	if td == nil {
		err = fmt.Errorf("trust domain does not exist: %q", tdName.String())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusNotFound)
	}

	response := api.TrustDomainFromEntity(td)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("trust domain entity - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// PutTrustDomainByName updates the trust domain - (PUT /trust-domain/{trustDomainName})
func (h *AdminAPIHandlers) PutTrustDomainByName(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	ctx := echoCtx.Request().Context()

	reqBody := &admin.PutTrustDomainByNameJSONRequestBody{}
	err := chttp.ParseRequestBodyToStruct(echoCtx, reqBody)
	if err != nil {
		err := fmt.Errorf("failed to read trust domain put body: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	etd, err := reqBody.ToEntity()
	if err != nil {
		err := fmt.Errorf("failed to read trust domain put body: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	_, err = h.lookupTrustDomain(ctx, trustDomainName)
	if err != nil {
		return err
	}

	td, err := h.Datastore.CreateOrUpdateTrustDomain(ctx, etd)
	if err != nil {
		err = fmt.Errorf("failed creating/updating trust domain: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	h.Logger.Printf("Trust Bundle %v updated", td.Name)

	response := api.TrustDomainFromEntity(td)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("relationships - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// UpdateRelationshipByID updates a specific relationship based on its id - (PATCH /relationships/{relationshipID})
func (h *AdminAPIHandlers) UpdateRelationshipByID(echoCtx echo.Context, relationshipID api.UUID) error {
	ctx := echoCtx.Request().Context()

	reqBody := &admin.UpdateRelationshipByIDJSONRequestBody{}
	err := chttp.ParseRequestBodyToStruct(echoCtx, reqBody)
	if err != nil {
		err := fmt.Errorf("failed to parse relationship patch body: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	rel, err := reqBody.ToEntity()
	if err != nil {
		err := fmt.Errorf("failed to read relationship patch body: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	relDB, err := h.findRelationshipByID(ctx, relationshipID)
	if err != nil {
		return err
	}

	// Set the ID to perform an Update instead of a Creation of a new Relationship.
	rel.ID = relDB.ID

	// Check if the user its performing an update in a single consent status, and replace the other one with the existing consent in the databse.
	if rel.TrustDomainAConsent == "" {
		rel.TrustDomainAConsent = relDB.TrustDomainAConsent
	}
	if rel.TrustDomainBConsent == "" {
		rel.TrustDomainBConsent = relDB.TrustDomainBConsent
	}

	relationship, err := h.Datastore.CreateOrUpdateRelationship(ctx, rel)
	if err != nil {
		err = fmt.Errorf("failed updating relationship: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	h.Logger.Printf("Relationship updated.")

	response := api.RelationshipFromEntity(relationship)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("relationship entity - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// DeleteRelationshipByID deletes a specific relationship based on its id - (DELETE /relationships/{relationshipID})
func (h *AdminAPIHandlers) DeleteRelationshipByID(echoCtx echo.Context, relationshipID api.UUID) error {
	ctx := echoCtx.Request().Context()

	err := h.Datastore.DeleteRelationship(ctx, relationshipID)
	if err != nil {
		err = fmt.Errorf("failed getting relationships: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	err = chttp.WriteResponse(echoCtx, http.StatusOK, "relationship deleted")
	if err != nil {
		err = fmt.Errorf("relationship entity - %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// GetJoinToken generates a join token for the trust domain - (GET /trust-domain/{trustDomainName}/join-token)
func (h *AdminAPIHandlers) GetJoinToken(echoCtx echo.Context, trustDomainName api.TrustDomainName, params admin.GetJoinTokenParams) error {
	ctx := echoCtx.Request().Context()
	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		err = fmt.Errorf("failed parsing trust domain: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		err = fmt.Errorf("failed looking up trust domain: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	if td == nil {
		err = fmt.Errorf("trust domain does not exist: %q", trustDomainName)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	token := uuid.New()

	ttl := time.Duration(params.Ttl) * time.Second
	joinToken := &entity.JoinToken{
		Token:         token.String(),
		TrustDomainID: td.ID.UUID,
		ExpiresAt:     time.Now().Add(ttl),
	}

	_, err = h.Datastore.CreateJoinToken(ctx, joinToken)
	if err != nil {
		err = fmt.Errorf("failed creating join token: %v", err)
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	response := admin.JoinTokenResponse{
		Token: token,
	}
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("failed to write join token response: %v", err.Error())
		return chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	h.Logger.WithField(telemetry.TrustDomain, tdName.String()).Debug("Created join token")

	return nil
}

func (h *AdminAPIHandlers) findTrustDomainByName(ctx context.Context, trustDomain string) (*entity.TrustDomain, error) {
	tdName, err := spiffeid.TrustDomainFromString(trustDomain)
	if err != nil {
		err = fmt.Errorf("failed parsing trust domain name: %v", err)
		return nil, chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		err = fmt.Errorf("failed getting trust domain: %v", err)
		return nil, chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return td, nil
}

func (h *AdminAPIHandlers) lookupTrustDomain(ctx context.Context, trustDomainName api.TrustDomainName) (*entity.TrustDomain, error) {
	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		return nil, fmt.Errorf("malformed trust domain[%q]: %v", trustDomainName, err)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		msg := "error looking up trust domain"
		err := fmt.Errorf("%s: %v", msg, err)
		return nil, chttp.LogAndRespondWithError(h.Logger, err, msg, http.StatusInternalServerError)
	}

	if td == nil {
		errMsg := fmt.Errorf("trust domain does not exist: %q", tdName.String())
		return nil, chttp.LogAndRespondWithError(h.Logger, errMsg, errMsg.Error(), http.StatusNotFound)
	}

	return td, nil
}

func (h *AdminAPIHandlers) findRelationshipByID(ctx context.Context, relationshipID api.UUID) (*entity.Relationship, error) {
	relationship, err := h.Datastore.FindRelationshipByID(ctx, relationshipID)
	if err != nil {
		err = fmt.Errorf("failed getting relationship: %v", err)
		return nil, chttp.LogAndRespondWithError(h.Logger, err, err.Error(), http.StatusInternalServerError)
	}

	return relationship, nil
}
