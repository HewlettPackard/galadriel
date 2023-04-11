package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
	"github.com/spiffe/go-spiffe/v2/spiffeid"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
)

func (e *Endpoints) createTrustDomainHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var trustDomainReq api.TrustDomain
	err := fromJSBody(r, &trustDomainReq)
	if err != nil {
		e.handleError(w, err.Error())
		return
	}

	// We may want to do some stuff before translating
	// So, thats why not encapsulating translate in json parsing
	dbTD, err := translate(trustDomainReq)
	if err != nil {
		e.handleError(w, err.Error())
		return
	}

	td, err := e.Datastore.FindTrustDomainByName(ctx, dbTD.Name)
	if err != nil {
		errMsg := fmt.Sprintf("failed looking up trust domain: %v", err)
		e.handleError(w, errMsg)
		return
	}
	if td != nil {
		errMsg := fmt.Sprintf("trust domain already exists: %q", trustDomainReq.Name)
		e.handleError(w, errMsg)
		return
	}

	m, err := e.Datastore.CreateOrUpdateTrustDomain(ctx, dbTD)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating trustDomain: %v", err)
		e.handleError(w, errMsg)
		return
	}

	e.Logger.Printf("Created trustDomain for trust domain: %s", trustDomainReq.Name)

	err = writeAsJSInResponse(w, m)
	if err != nil {
		errMsg := fmt.Sprintf("trustDomain entity - %v", err.Error())
		e.handleError(w, errMsg)
		return
	}
}

func (e *Endpoints) listTrustDomainsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ms, err := e.Datastore.ListTrustDomains(ctx)
	if err != nil {
		errMsg := fmt.Sprintf("failed listing trustDomains: %v", err)
		e.handleError(w, errMsg)
		return
	}

	err = writeAsJSInResponse(w, ms)
	if err != nil {
		errMsg := fmt.Sprintf("trustDomains - %v", err.Error())
		e.handleError(w, errMsg)
		return
	}
}

func (e *Endpoints) createRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed reading request body: %v", err)
		e.handleError(w, errMsg)
		return
	}

	var relationshipReq entity.Relationship
	if err = json.Unmarshal(body, &relationshipReq); err != nil {
		errMsg := fmt.Sprintf("failed unmarshalling request: %v", err)
		e.handleError(w, errMsg)
		return
	}

	tdaID, err := e.Datastore.FindTrustDomainByName(ctx, relationshipReq.TrustDomainAName)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating relationship: %v", err)
		e.handleError(w, errMsg)
		return
	}
	relationshipReq.TrustDomainAID = tdaID.ID.UUID

	tdbID, err := e.Datastore.FindTrustDomainByName(ctx, relationshipReq.TrustDomainBName)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating relationship: %v", err)
		e.handleError(w, errMsg)
		return
	}
	relationshipReq.TrustDomainBID = tdbID.ID.UUID

	rel, err := e.Datastore.CreateOrUpdateRelationship(ctx, &relationshipReq)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating relationship: %v", err)
		e.handleError(w, errMsg)
		return
	}

	e.Logger.Printf("Created relationship between trust domains %s and %s", rel.TrustDomainAID, rel.TrustDomainBID)

	err = writeAsJSInResponse(w, rel)
	if err != nil {
		errMsg := fmt.Sprintf("relationships - %v", err.Error())
		e.handleError(w, errMsg)
		return
	}
}

func (e *Endpoints) listRelationshipsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rels, err := e.Datastore.ListRelationships(context.TODO())
	if err != nil {
		errMsg := fmt.Sprintf("failed listing relationships: %v", err)
		e.handleError(w, errMsg)
		return
	}

	rels, err = e.populateTrustDomainNames(ctx, rels)
	if err != nil {
		errMsg := fmt.Sprintf("failed populating relationships entities: %v", err)
		e.handleError(w, errMsg)
		return
	}

	err = writeAsJSInResponse(w, rels)
	if err != nil {
		errMsg := fmt.Sprintf("relationships entities - %v", err.Error())
		e.handleError(w, errMsg)
		return
	}
}

func (e *Endpoints) populateTrustDomainNames(ctx context.Context, relationships []*entity.Relationship) ([]*entity.Relationship, error) {
	for _, r := range relationships {
		tda, err := e.Datastore.FindTrustDomainByID(ctx, r.TrustDomainAID)
		if err != nil {
			return nil, err
		}
		r.TrustDomainAName = tda.Name

		tdb, err := e.Datastore.FindTrustDomainByID(ctx, r.TrustDomainBID)
		if err != nil {
			return nil, err
		}
		r.TrustDomainBName = tdb.Name
	}
	return relationships, nil
}

func (e *Endpoints) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed reading request body: %v", err)
		e.handleError(w, errMsg)
		return
	}

	var trustDomain api.TrustDomain
	if err = json.Unmarshal(body, &trustDomain); err != nil {
		errMsg := fmt.Sprintf("failed unmarshalling request: %v", err)
		e.handleError(w, errMsg)
		return
	}

	tdName, err := spiffeid.TrustDomainFromString(trustDomain.Name)
	if err != nil {
		errMsg := fmt.Sprintf("malformed trust domain name: %v", err)
		e.handleError(w, errMsg)
		return
	}

	tdID, err := e.Datastore.FindTrustDomainByName(ctx, tdName)
	if err != nil {
		errMsg := fmt.Sprintf("could not find trust domain name: %v", err)
		e.handleError(w, errMsg)
		return
	}

	token, err := util.GenerateToken()
	if err != nil {
		errMsg := fmt.Sprintf("failed generating token: %v", err)
		e.handleError(w, errMsg)
		return
	}

	newToken := &entity.JoinToken{
		TrustDomainID: tdID.ID.UUID,
		Token:         token,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	at, err := e.Datastore.CreateJoinToken(ctx, newToken)

	if err != nil {
		errMsg := fmt.Sprintf("failed generating access token: %v", err)
		e.handleError(w, errMsg)
		return
	}

	err = writeAsJSInResponse(w, at)
	if err != nil {
		errMsg := fmt.Sprintf("access token - %v", err.Error())
		e.handleError(w, errMsg)
		return
	}
}

func (e *Endpoints) onboardHandler(c echo.Context) error {
	e.Logger.Info("Harvester connected")
	return nil
}

func (e *Endpoints) validateToken(ctx echo.Context, token string) (bool, error) {
	t, err := e.Datastore.FindJoinToken(ctx.Request().Context(), token)
	if err != nil {
		e.Logger.Errorf("Invalid Token: %s\n", token)
		return false, err
	}

	e.Logger.Debugf("Token valid for trust domain: %s\n", t.TrustDomainID)

	ctx.Set("token", t)

	return true, nil
}

/*
fromJSBody parses json bytes into a struct
r: Request that contains the json bytes to be parsed into 'in'
in: Reference(pointer) to the interface to be full filled
*/
func fromJSBody(r *http.Request, in interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed reading request body: %v", err)
	}

	if err = json.Unmarshal(body, in); err != nil {
		return fmt.Errorf("failed unmarshalling request: %v", err)
	}

	return nil
}

/*
writeAsJSInResponse parses a struct into a json and writes in the response
w: The response writer to be full filled with the struct response bytes
out: A pointer to the interface to be writed in the response
*/
func writeAsJSInResponse(w http.ResponseWriter, out interface{}) error {
	outBytes, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("failed marshalling : %v", err)
	}

	_, err = w.Write(outBytes)
	if err != nil {
		return fmt.Errorf("failed writing response: %v", err)
	}

	return nil
}

/*
translate maps the API models into the Database models performing some check guarantees
td: The api Trust Domain structure
*/
func translate(td api.TrustDomain) (*entity.TrustDomain, error) {
	harvesterSpiffeID, err := spiffeid.FromString(*td.HarvesterSpiffeId)
	if err != nil {
		return nil, ErrWrongSPIFFEID{cause: err}
	}

	tdName, err := spiffeid.TrustDomainFromString(td.Name)
	if err != nil {
		return nil, ErrWrongTrustDomain{cause: err}
	}

	description := ""
	if td.Description != nil {
		description = *td.Description
	}

	onboardingBundle := []byte{}
	if td.OnboardingBundle != nil {
		onboardingBundle = []byte(*td.OnboardingBundle)
	}

	uuid := uuid.NullUUID{
		UUID:  td.Id,
		Valid: true,
	}

	return &entity.TrustDomain{
		ID:                uuid,
		Name:              tdName,
		CreatedAt:         td.CreatedAt,
		UpdatedAt:         td.UpdatedAt,
		Description:       description,
		OnboardingBundle:  onboardingBundle,
		HarvesterSpiffeID: harvesterSpiffeID,
	}, nil
}

func (e *Endpoints) handleError(w http.ResponseWriter, errMsg string) {
	errMsg = util.LogSanitize(errMsg)
	e.Logger.Errorf(errMsg)

	errBytes := []byte(errMsg)
	w.WriteHeader(500)
	_, err := w.Write(errBytes)
	if err != nil {
		e.Logger.Errorf("Failed to write error response: %v", err)
	}

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true)
	err = encoder.Encode(errBytes)
	if err != nil {
		e.Logger.Errorf("Failed to write error response: %v", err)
	}
}
