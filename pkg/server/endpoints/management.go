package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/labstack/echo/v4"

	"github.com/HewlettPackard/galadriel/pkg/common/util"
)

func (e *Endpoints) createTrustDomainHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed reading request body: %v", err)
		e.handleError(w, errMsg)
		return
	}

	var trustDomainReq entity.TrustDomain
	if err = json.Unmarshal(body, &trustDomainReq); err != nil {
		errMsg := fmt.Sprintf("failed unmarshalling request: %v", err)
		e.handleError(w, errMsg)
		return
	}

	td, err := e.Datastore.FindTrustDomainByName(ctx, trustDomainReq.Name)
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

	m, err := e.Datastore.CreateOrUpdateTrustDomain(ctx, &trustDomainReq)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating trustDomain: %v", err)
		e.handleError(w, errMsg)
		return
	}

	e.Logger.Printf("Created trustDomain for trust domain: %s", trustDomainReq.Name)

	trustDomainBytes, err := json.Marshal(m)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling trustDomain entity: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(trustDomainBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
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

	trustDomainsBytes, err := json.Marshal(ms)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling trustDomains entities: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(trustDomainsBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
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

	relBytes, err := json.Marshal(rel)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling trustDomainship entity: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(relBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
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

	relsBytes, err := json.Marshal(rels)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling relationships entities: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(relsBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
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

	var trustDomain entity.TrustDomain
	if err = json.Unmarshal(body, &trustDomain); err != nil {
		errMsg := fmt.Sprintf("failed unmarshalling request: %v", err)
		e.handleError(w, errMsg)
		return
	}

	tdID, err := e.Datastore.FindTrustDomainByName(ctx, trustDomain.Name)
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

	tokenBytes, err := json.Marshal(at)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling access token entity: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(tokenBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
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
