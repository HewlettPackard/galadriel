package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
)

func (e *Endpoints) createMemberHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed reading request body: %v", err)
		e.handleError(w, errMsg)
		return
	}

	var memberReq common.Member
	if err = json.Unmarshal(body, &memberReq); err != nil {
		errMsg := fmt.Sprintf("failed unmarshalling request: %v", err)
		e.handleError(w, errMsg)
		return
	}

	m, err := e.DataStore.CreateMember(context.TODO(), &memberReq)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating member: %v", err)
		e.handleError(w, errMsg)
		return
	}

	e.Log.Printf("Created member for trust domain: %s", memberReq.TrustDomain)

	memberBytes, err := json.Marshal(m)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling member entity: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(memberBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
		e.handleError(w, errMsg)
		return
	}

}

func (e *Endpoints) listMembersHandler(w http.ResponseWriter, r *http.Request) {
	ms, err := e.DataStore.ListMembers(context.TODO())
	if err != nil {
		errMsg := fmt.Sprintf("failed listing members: %v", err)
		e.handleError(w, errMsg)
		return
	}

	e.Log.Println("Members: %d", len(ms))

	membersBytes, err := json.Marshal(ms)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling members entities: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(membersBytes)
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
		e.handleError(w, errMsg)
		return
	}

}

func (e *Endpoints) createRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed reading request body: %v", err)
		e.handleError(w, errMsg)
		return
	}

	var relationshipReq common.Relationship
	if err = json.Unmarshal(body, &relationshipReq); err != nil {
		errMsg := fmt.Sprintf("failed unmarshalling request: %v", err)
		e.handleError(w, errMsg)
		return
	}

	rel, err := e.DataStore.CreateRelationship(context.TODO(), &relationshipReq)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating relationship: %v", err)
		e.handleError(w, errMsg)
		return
	}

	e.Log.Printf("Created relationship between trust domains %s and %s", relationshipReq.TrustDomainA, relationshipReq.TrustDomainB)

	relBytes, err := json.Marshal(rel)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling membership entity: %v", err)
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
	ms, err := e.DataStore.ListRelationships(context.TODO())
	if err != nil {
		errMsg := fmt.Sprintf("failed listing relationships: %v", err)
		e.handleError(w, errMsg)
		return
	}

	relsBytes, err := json.Marshal(ms)
	if err != nil {
		errMsg := fmt.Sprintf("failed marshalling relationships entities: %v", err)
		e.handleError(w, errMsg)
		return
	}

	_, err = w.Write(relsBytes)
	e.Log.Infof("rels len: %d", len(relsBytes))
	e.Log.Info(string(relsBytes))
	if err != nil {
		errMsg := fmt.Sprintf("failed writing response: %v", err)
		e.handleError(w, errMsg)
		return
	}
}

func (e *Endpoints) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed reading request body: %v", err)
		e.handleError(w, errMsg)
		return
	}

	var member common.Member
	if err = json.Unmarshal(body, &member); err != nil {
		errMsg := fmt.Sprintf("failed unmarshalling request: %v", err)
		e.handleError(w, errMsg)
		return
	}

	token, err := util.GenerateToken()
	if err != nil {
		errMsg := fmt.Sprintf("failed generating token: %v", err)
		e.handleError(w, errMsg)
		return
	}

	at, err := e.DataStore.GenerateAccessToken(
		context.TODO(), &common.AccessToken{Token: token, MemberID: member.ID, Expiry: time.Now().Add(1 * time.Hour)}, member.TrustDomain.String())
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
	e.Log.Info("Harvester connected")
	return nil
}

func (e *Endpoints) validateToken(token string) (bool, error) {
	t, err := e.DataStore.FetchAccessToken(context.TODO(), token)
	if err != nil {
		e.Log.Errorf("Invalid Token: %s\n", token)
		return false, err
	}

	e.Log.Infof("Token valid for trust domain: %s\n", t.Member.TrustDomain)

	return true, nil
}

func (e *Endpoints) handleError(w http.ResponseWriter, errMsg string) {
	e.Log.Errorf(errMsg)
	w.WriteHeader(500)
	_, err := w.Write([]byte(errMsg))
	if err != nil {
		e.Log.Errorf("Failed to write error response: %v", err)
	}
}
