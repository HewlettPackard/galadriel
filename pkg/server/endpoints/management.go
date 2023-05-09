package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	chttp "github.com/HewlettPackard/galadriel/pkg/common/http"
	"github.com/HewlettPackard/galadriel/pkg/server/api/admin"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"

	"github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
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
func (h *AdminAPIHandlers) GetRelationships(ctx echo.Context, params admin.GetRelationshipsParams) error {
	gctx := ctx.Request().Context()

	var err error
	var rels []*entity.Relationship

	if params.TrustDomainName != nil {
		td, err := h.findTrustDomainByName(gctx, *params.TrustDomainName)
		if err != nil {
			err = fmt.Errorf("failed parsing trust domain name: %v", err)
			return h.handleAndLog(err, http.StatusBadRequest)
		}

		rels, err = h.Datastore.FindRelationshipsByTrustDomainID(gctx, td.ID.UUID)
		if err != nil {
			err = fmt.Errorf("failed listing relationships: %v", err)
			return h.handleAndLog(err, http.StatusInternalServerError)
		}
	} else {
		rels, err = h.Datastore.ListRelationships(gctx)
		if err != nil {
			err = fmt.Errorf("failed listing relationships: %v", err)
			return h.handleAndLog(err, http.StatusInternalServerError)
		}
	}

	rels, err = h.filterRelationshipsByStatus(gctx, rels, params.Status)
	if err != nil {
		return err
	}

	rels, err = h.populateTrustDomainNames(gctx, rels)
	if err != nil {
		err = fmt.Errorf("failed populating relationships entities: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	cRelationships := mapRelationships(rels)
	err = chttp.WriteResponse(ctx, cRelationships)
	if err != nil {
		err = fmt.Errorf("relationships entities - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PutRelationships create a new relationship request between two trust domains - (PUT /relationships)
func (h AdminAPIHandlers) PutRelationships(ctx echo.Context) error {
	gctx := ctx.Request().Context()

	reqBody := &admin.PutRelationshipsJSONRequestBody{}
	err := chttp.FromBody(ctx, reqBody)
	if err != nil {
		err := fmt.Errorf("failed to read relationship put body: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	// Possible scenario when a fake trust domain uuid is used will fail to create
	// a relationship and a bad request should be raised.
	// Should we query the trust domains before trying to create a relation ??
	eRelationship := reqBody.ToEntity()
	rel, err := h.Datastore.CreateOrUpdateRelationship(gctx, eRelationship)
	if err != nil {
		err = fmt.Errorf("failed creating relationship: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Printf("Created relationship between trust domains %s and %s", rel.TrustDomainAID, rel.TrustDomainBID)

	response := api.RelationshipFromEntity(rel)
	err = chttp.WriteResponse(ctx, response)
	if err != nil {
		err = fmt.Errorf("relationships - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// GetRelationshipsRelationshipID retrieve a specific relationship based on its id - (GET /relationships/{relationshipID})
func (h AdminAPIHandlers) GetRelationshipsRelationshipID(ctx echo.Context, relationshipID api.UUID) error {
	gctx := ctx.Request().Context()

	r, err := h.Datastore.FindRelationshipByID(gctx, relationshipID)
	if err != nil {
		err = fmt.Errorf("failed getting relationships: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	response := api.RelationshipFromEntity(r)
	err = chttp.WriteResponse(ctx, response)
	if err != nil {
		err = fmt.Errorf("relationship entity - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PutTrustDomain create a new trust domain - (PUT /trust-domain)
func (h AdminAPIHandlers) PutTrustDomain(ctx echo.Context) error {
	// Getting golang context
	gctx := ctx.Request().Context()

	reqBody := &admin.PutTrustDomainJSONRequestBody{}
	err := chttp.FromBody(ctx, reqBody)
	if err != nil {
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	dbTD, err := reqBody.ToEntity()
	if err != nil {
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(gctx, dbTD.Name)
	if err != nil {
		err = fmt.Errorf("failed looking up trust domain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}
	if td != nil {
		err = fmt.Errorf("trust domain already exists: %q", dbTD.Name)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	m, err := h.Datastore.CreateOrUpdateTrustDomain(gctx, dbTD)
	if err != nil {
		err = fmt.Errorf("failed creating trustDomain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Printf("Created trustDomain for trust domain: %s", dbTD.Name)

	response := api.TrustDomainFromEntity(m)
	err = chttp.WriteResponse(ctx, response)
	if err != nil {
		err = fmt.Errorf("trustDomain entity - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// GetTrustDomainTrustDomainName retrieve a specific trust domain by its name - (GET /trust-domain/{trustDomainName})
func (h AdminAPIHandlers) GetTrustDomainTrustDomainName(ctx echo.Context, trustDomainName api.TrustDomainName) error {
	gctx := ctx.Request().Context()

	tdName, err := spiffeid.TrustDomainFromString(trustDomainName)
	if err != nil {
		err = fmt.Errorf("failed parsing trust domain name: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	td, err := h.Datastore.FindTrustDomainByName(gctx, tdName)
	if err != nil {
		err = fmt.Errorf("failed getting trust domain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	response := api.TrustDomainFromEntity(td)
	err = chttp.WriteResponse(ctx, response)
	if err != nil {
		err = fmt.Errorf("trust domain entity - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PutTrustDomainTrustDomainName updates the trust domain - (PUT /trust-domain/{trustDomainName})
func (h AdminAPIHandlers) PutTrustDomainTrustDomainName(ctx echo.Context, trustDomainName api.UUID) error {
	gctx := ctx.Request().Context()

	reqBody := &admin.PutTrustDomainTrustDomainNameJSONRequestBody{}
	err := chttp.FromBody(ctx, reqBody)
	if err != nil {
		err := fmt.Errorf("failed to read trust domain put body: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	etd, err := reqBody.ToEntity()
	if err != nil {
		err := fmt.Errorf("failed to read trust domain put body: %v", err)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	td, err := h.Datastore.CreateOrUpdateTrustDomain(gctx, etd)
	if err != nil {
		err = fmt.Errorf("failed creating/updating trust domain: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Printf("Trust Bundle %v created/updated", td.Name)

	response := api.TrustDomainFromEntity(td)
	err = chttp.WriteResponse(ctx, response)
	if err != nil {
		err = fmt.Errorf("relationships - %v", err.Error())
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	return nil
}

// PostTrustDomainTrustDomainNameJoinToken generate a join token for the trust domain - (POST /trust-domain/{trustDomainName}/join-token)
func (h AdminAPIHandlers) PostTrustDomainTrustDomainNameJoinToken(ctx echo.Context, trustDomainName api.TrustDomainName) error {
	gctx := ctx.Request().Context()

	td, err := h.findTrustDomainByName(gctx, trustDomainName)
	if err != nil {
		return err
	}

	if td == nil {
		err = fmt.Errorf("trust domain does not exists %v", trustDomainName)
		return h.handleAndLog(err, http.StatusBadRequest)
	}

	token, err := util.GenerateToken()
	if err != nil {
		err = fmt.Errorf("failed generating a new join token %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	jt := &entity.JoinToken{
		TrustDomainID: td.ID.UUID,
		Token:         token,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	jt, err = h.Datastore.CreateJoinToken(gctx, jt)
	if err != nil {
		err = fmt.Errorf("failed creating the join token: %v", err)
		return h.handleAndLog(err, http.StatusInternalServerError)
	}

	h.Logger.Printf("join token successfully created for %v", td.Name)

	response := admin.JoinTokenResult{
		Token: uuid.MustParse(jt.Token),
	}

	err = chttp.WriteResponse(ctx, response)
	if err != nil {
		err = fmt.Errorf("relationships - %v", err.Error())
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
