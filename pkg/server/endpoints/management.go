// TODO: rename this file to admin.go
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
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
)

type AdminAPIHandlers struct {
	Logger    logrus.FieldLogger
	Datastore datastore.Datastore
}

// NewAdminAPIHandlers create a new NewAdminAPIHandlers
func NewAdminAPIHandlers(l logrus.FieldLogger, ds datastore.Datastore) *AdminAPIHandlers {
	return &AdminAPIHandlers{
		Logger:    l,
		Datastore: ds,
	}
}

// GetRelationships list all relationships filtered by the request params - (GET /relationships)
func (h *AdminAPIHandlers) GetRelationships(echoContext echo.Context, params admin.GetRelationshipsParams) error {
	ctx := echoContext.Request().Context()

	var err error
	var rels []*entity.Relationship

	if params.TrustDomainName != nil {
		td, err := h.findTrustDomainByName(ctx, *params.TrustDomainName)
		if err != nil {
			err = fmt.Errorf("failed looking up trust domain name: %w", err)
			return h.handleAndLog(err, http.StatusBadRequest)
		}

		rels, err = h.Datastore.FindRelationshipsByTrustDomainID(ctx, td.ID.UUID)
		if err != nil {
			err = fmt.Errorf("failed looking up relationships: %v", err)
			return h.handleAndLog(err, http.StatusInternalServerError)
		}
	} else {
		rels, err = h.Datastore.ListRelationships(ctx)
		if err != nil {
			err = fmt.Errorf("failed listing relationships: %v", err)
			return h.handleAndLog(err, http.StatusInternalServerError)
		}
	}

	rels, err = h.filterRelationshipsByStatus(ctx, rels, params.Status)
	if err != nil {
		return err
	}

	rels, err = h.populateTrustDomainNames(ctx, rels)
	if err != nil {
		err = fmt.Errorf("failed populating relationships entities: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	cRelationships := mapRelationships(rels)
	err = chttp.WriteResponse(echoContext, http.StatusOK, cRelationships)
	if err != nil {
		err = fmt.Errorf("relationships entities - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PutRelationships create a new relationship request between two trust domains - (PUT /relationships)
func (h *AdminAPIHandlers) PutRelationship(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	reqBody := &admin.PutRelationshipJSONRequestBody{}
	err := chttp.FromBody(echoCtx, reqBody)
	if err != nil {
		err := fmt.Errorf("failed to read relationship put body: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}
	eRelationship := reqBody.ToEntity()

	_, err = h.lookupTrustDomain(ctx, eRelationship.TrustDomainAID)
	if err != nil {
		return err
	}

	_, err = h.lookupTrustDomain(ctx, eRelationship.TrustDomainBID)
	if err != nil {
		return err
	}

	rel, err := h.Datastore.CreateOrUpdateRelationship(ctx, eRelationship)
	if err != nil {
		err = fmt.Errorf("failed creating relationship: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Printf("Created relationship between trust domains %s and %s", rel.TrustDomainAID, rel.TrustDomainBID)

	response := api.RelationshipFromEntity(rel)
	err = chttp.WriteResponse(echoCtx, http.StatusCreated, response)
	if err != nil {
		err = fmt.Errorf("relationships - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// GetRelationshipByID retrieve a specific relationship based on its id - (GET /relationships/{relationshipID})
func (h *AdminAPIHandlers) GetRelationshipByID(echoCtx echo.Context, relationshipID api.UUID) error {
	ctx := echoCtx.Request().Context()

	r, err := h.Datastore.FindRelationshipByID(ctx, relationshipID)
	if err != nil {
		err = fmt.Errorf("failed getting relationships: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	if r == nil {
		err = errors.New("relationship not found")
		return h.handleAndLog(err, http.StatusNotFound)
	}

	response := api.RelationshipFromEntity(r)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("relationship entity - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PutTrustDomain create a new trust domain - (PUT /trust-domain)
func (h *AdminAPIHandlers) PutTrustDomain(echoCtx echo.Context) error {
	// Getting golang context
	ctx := echoCtx.Request().Context()

	reqBody := &admin.PutTrustDomainJSONRequestBody{}
	err := chttp.FromBody(echoCtx, reqBody)
	if err != nil {
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	dbTD, err := reqBody.ToEntity()
	if err != nil {
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, dbTD.Name)
	if err != nil {
		err = fmt.Errorf("failed looking up trust domain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	if td != nil {
		err = fmt.Errorf("trust domain already exists: %q", dbTD.Name)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	m, err := h.Datastore.CreateOrUpdateTrustDomain(ctx, dbTD)
	if err != nil {
		err = fmt.Errorf("failed creating trustDomain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Printf("Created trustDomain for trust domain: %s", dbTD.Name)

	response := api.TrustDomainFromEntity(m)
	err = chttp.WriteResponse(echoCtx, http.StatusCreated, response)
	if err != nil {
		err = fmt.Errorf("trustDomain entity - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// TODO: customize these names.
// GetTrustDomainByName retrieve a specific trust domain by its name - (GET /trust-domain/{trustDomainName})
func (h *AdminAPIHandlers) GetTrustDomainByName(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	ctx := echoCtx.Request().Context()

	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		err = fmt.Errorf("failed parsing trust domain name: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		err = fmt.Errorf("failed getting trust domain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	if td == nil {
		err = fmt.Errorf("trust domain does not exists")
		return h.handleAndLog(err, http.StatusNotFound)
	}

	response := api.TrustDomainFromEntity(td)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("trust domain entity - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PutTrustDomainTrustDomainName updates the trust domain - (PUT /trust-domain/{trustDomainName})
func (h *AdminAPIHandlers) PutTrustDomainByName(echoCtx echo.Context, trustDomainID api.UUID) error {
	ctx := echoCtx.Request().Context()

	reqBody := &admin.PutTrustDomainByNameJSONRequestBody{}
	err := chttp.FromBody(echoCtx, reqBody)
	if err != nil {
		err := fmt.Errorf("failed to read trust domain put body: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	etd, err := reqBody.ToEntity()
	if err != nil {
		err := fmt.Errorf("failed to read trust domain put body: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	_, err = h.lookupTrustDomain(ctx, trustDomainID)
	if err != nil {
		return err
	}

	td, err := h.Datastore.CreateOrUpdateTrustDomain(ctx, etd)
	if err != nil {
		err = fmt.Errorf("failed creating/updating trust domain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Printf("Trust Bundle %v updated", td.Name)

	response := api.TrustDomainFromEntity(td)
	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("relationships - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// GetJoinToken generate a join token for the trust domain - (GET /trust-domain/{trustDomainName}/join-token)
func (h *AdminAPIHandlers) GetJoinToken(echoCtx echo.Context, trustDomainName api.TrustDomainName) error {
	ctx := echoCtx.Request().Context()
	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		err = fmt.Errorf("failed parsing trust domain: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		err = fmt.Errorf("failed looking up trust domain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	if td == nil {
		errMsg := fmt.Errorf("trust domain exists: %q", trustDomainName)
		return h.handleAndLog(errMsg, http.StatusBadRequest)
	}

	token := uuid.New()

	joinToken := &entity.JoinToken{
		Token:         token.String(),
		TrustDomainID: td.ID.UUID,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	_, err = h.Datastore.CreateJoinToken(ctx, joinToken)
	if err != nil {
		err = fmt.Errorf("failed creating join token: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Infof("Created join token for trust domain: %s", tdName)

	response := admin.JoinTokenResult{
		Token: token,
	}

	err = chttp.WriteResponse(echoCtx, http.StatusOK, response)
	if err != nil {
		err = fmt.Errorf("failed to write join token response: %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

func (h *AdminAPIHandlers) findTrustDomainByName(ctx context.Context, trustDomain string) (*entity.TrustDomain, error) {
	tdName, err := spiffeid.TrustDomainFromString(trustDomain)
	if err != nil {
		err = fmt.Errorf("failed parsing trust domain name: %v", err)
		return nil, h.handleAndLog(err, http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		err = fmt.Errorf("failed getting trust domain: %v", err)
		return nil, h.handleAndLog(err, http.StatusInternalServerError)
	}

	return td, nil
}

func (h *AdminAPIHandlers) filterRelationshipsByStatus(
	ctx context.Context,
	relationships []*entity.Relationship,
	status *admin.GetRelationshipsParamsStatus,
) ([]*entity.Relationship, error) {

	if status != nil {
		switch *status {
		case admin.Denied:
			return filterBy(relationships, deniedRelationFilter), nil
		case admin.Approved:
			return filterBy(relationships, approvedRelationFilter), nil
		case admin.Pending:
			return filterBy(relationships, pendingRelationFilter), nil
		}

		err := fmt.Errorf(
			"unrecognized status filter %v, accepted values [%v, %v, %v]",
			*status, admin.Denied, admin.Approved, admin.Pending,
		)
		return nil, h.handleAndLog(err, http.StatusBadRequest)
	} else {
		return filterBy(relationships, pendingRelationFilter), nil
	}
}

func (h *AdminAPIHandlers) populateTrustDomainNames(ctx context.Context, relationships []*entity.Relationship) ([]*entity.Relationship, error) {
	for _, r := range relationships {
		tda, err := h.Datastore.FindTrustDomainByID(ctx, r.TrustDomainAID)
		if err != nil {
			return nil, err
		}
		r.TrustDomainAName = tda.Name

		tdb, err := h.Datastore.FindTrustDomainByID(ctx, r.TrustDomainBID)
		if err != nil {
			return nil, err
		}
		r.TrustDomainBName = tdb.Name
	}
	return relationships, nil
}

func (h *AdminAPIHandlers) lookupTrustDomain(ctx context.Context, trustDomainID uuid.UUID) (*entity.TrustDomain, error) {
	td, err := h.Datastore.FindTrustDomainByID(ctx, trustDomainID)
	if err != nil {
		msg := errors.New("error looking up trust domain")
		errMsg := fmt.Errorf("%s: %w", msg, err)
		return nil, h.handleAndLog(errMsg, http.StatusInternalServerError)
	}

	if td == nil {
		errMsg := fmt.Errorf("trust domain exists: %q", trustDomainID)
		return nil, h.handleAndLog(errMsg, http.StatusBadRequest)
	}

	return td, nil
}

func mapRelationships(relationships []*entity.Relationship) []*api.Relationship {
	cRelationships := []*api.Relationship{}

	for _, r := range relationships {
		cRelation := api.RelationshipFromEntity(r)
		cRelationships = append(cRelationships, cRelation)
	}

	return cRelationships
}

func (h *AdminAPIHandlers) handleAndLog(err error, code int) error {
	errMsg := util.LogSanitize(err.Error())
	h.Logger.Errorf(errMsg)
	return echo.NewHTTPError(code, err.Error())
}

func deniedRelationFilter(e *entity.Relationship) bool {
	return !e.TrustDomainAConsent || !e.TrustDomainBConsent
}

func approvedRelationFilter(e *entity.Relationship) bool {
	return e.TrustDomainAConsent && e.TrustDomainBConsent
}

func pendingRelationFilter(e *entity.Relationship) bool {
	return !e.TrustDomainAConsent || !e.TrustDomainBConsent
}

// filterBy will generate a new slice with the elements that matched
func filterBy[E any](s []E, match func(E) bool) []E {
	filtered := []E{}
	for _, e := range s {
		if match(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
