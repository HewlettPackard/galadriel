package endpoints

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
)

func (e *Endpoints) createMemberHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.Log.Errorf("failed reading body: %v", err)
		w.WriteHeader(500)
		return
	}

	var memberReq common.Member
	if err = json.Unmarshal(body, &memberReq); err != nil {
		e.Log.Errorf("failed unmarshalling body: %v", err)
		w.WriteHeader(500)
		return
	}

	m, err := e.DataStore.CreateMember(context.TODO(), &memberReq)
	if err != nil {
		e.Log.Errorf("failed creating member: %v", err)
		w.WriteHeader(500)
		return
	}

	memberBytes, err := json.Marshal(m)
	if err != nil {
		e.Log.Errorf("failed marshalling member: %v", err)
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(memberBytes)
	if err != nil {
		return
	}

}

func (e *Endpoints) createRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.Log.Errorf("failed reading body: %v", err)
		w.WriteHeader(500)
		return
	}

	var relationshipReq common.Relationship
	if err = json.Unmarshal(body, &relationshipReq); err != nil {
		e.Log.Errorf("failed unmarshalling body: %v", err)
		w.WriteHeader(500)
		return
	}

	rel, err := e.DataStore.CreateRelationship(context.TODO(), &relationshipReq)
	if err != nil {
		e.Log.Errorf("failed creating relationship: %v", err)
		w.WriteHeader(500)
		return
	}

	relBytes, err := json.Marshal(rel)
	if err != nil {
		e.Log.Errorf("failed marshalling relationship: %v", err)
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(relBytes)
	if err != nil {
		return
	}
}

func (e *Endpoints) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.Log.Errorf("failed parsing body: %v", err)
		w.WriteHeader(500)
		return
	}

	var member common.Member
	if err = json.Unmarshal(body, &member); err != nil {
		e.Log.Errorf("failed unmarshalling body: %v", err)
		w.WriteHeader(500)
		return
	}

	token, err := util.GenerateToken()
	if err != nil {
		e.Log.Errorf("failed generating token: %v", err)
		w.WriteHeader(500)
		return
	}

	at, err := e.DataStore.GenerateAccessToken(
		context.TODO(), &common.AccessToken{Token: token, MemberID: member.ID, Expiry: time.Now()}, member.TrustDomain)
	if err != nil {
		e.Log.Errorf("failed creating access token: %v", err)
		w.WriteHeader(500)
		return
	}

	tokenBytes, err := json.Marshal(at)
	if err != nil {
		e.Log.Errorf("failed marshalling token: %v", err)
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(tokenBytes)
	if err != nil {
		e.Log.Errorf("failed to return token: %v", err)
		w.WriteHeader(500)
		return
	}
}
