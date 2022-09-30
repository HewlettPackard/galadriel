package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
)

func (e *Endpoints) createMemberHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.handleError(w, err)
		return
	}

	var memberReq common.Member
	if err = json.Unmarshal(body, &memberReq); err != nil {
		e.handleError(w, err)
		return
	}

	m, err := e.DataStore.CreateMember(context.TODO(), &memberReq)
	if err != nil {
		errMsg := fmt.Sprintf("failed creating member: %v", err)
		e.Log.Errorf(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	e.Log.Printf("Created member for trust domain: %s", memberReq.TrustDomain)

	memberBytes, err := json.Marshal(m)
	if err != nil {
		e.handleError(w, err)
		return
	}

	_, err = w.Write(memberBytes)
	if err != nil {
		e.handleError(w, err)
		return
	}

}

func (e *Endpoints) createRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.handleError(w, err)
		return
	}

	var relationshipReq common.Relationship
	if err = json.Unmarshal(body, &relationshipReq); err != nil {
		e.handleError(w, err)
		return
	}

	rel, err := e.DataStore.CreateRelationship(context.TODO(), &relationshipReq)
	if err != nil {
		e.Log.Errorf("failed creating relationship: %v", err)
		w.WriteHeader(500)
		return
	}

	e.Log.Printf("Created relationship between trust domains %s and %s", relationshipReq.TrustDomainA, relationshipReq.TrustDomainB)

	relBytes, err := json.Marshal(rel)
	if err != nil {
		e.handleError(w, err)
		return
	}

	_, err = w.Write(relBytes)
	if err != nil {
		e.handleError(w, err)
		return
	}
}

func (e *Endpoints) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		e.handleError(w, err)
		return
	}

	var member common.Member
	if err = json.Unmarshal(body, &member); err != nil {
		e.handleError(w, err)
		return
	}

	token, err := util.GenerateToken()
	if err != nil {
		e.handleError(w, err)
		return
	}

	at, err := e.DataStore.GenerateAccessToken(
		context.TODO(), &common.AccessToken{Token: token, MemberID: member.ID, Expiry: time.Now().Add(1 * time.Hour)}, member.TrustDomain.IDString())
	if err != nil {
		errMsg := fmt.Sprintf("failed creating access token: %v", err)
		e.Log.Error(errMsg)
		w.Write([]byte(errMsg))
		return
	}

	tokenBytes, err := json.Marshal(at)
	if err != nil {
		e.handleError(w, err)
		return
	}

	_, err = w.Write(tokenBytes)
	if err != nil {
		e.handleError(w, err)
		return
	}
}

func (e *Endpoints) onboardHandler(c echo.Context) error {
	header := c.Request().Header.Get("Authorization")
	err := e.validateToken(header)
	if err != nil {
		return c.String(400, "Invalid token")
	}

	e.Log.Info("Harvester connected")
	return nil
}

func (e *Endpoints) validateToken(header string) error {
	splitToken := strings.Split(header, "Bearer ")
	if len(splitToken) != 2 {
		return errors.New("invalid token")
	}
	token := splitToken[1]

	t, err := e.DataStore.FetchAccessToken(context.TODO(), token)
	if err != nil {
		e.Log.Errorf("Invalid Token: %s\n", token)
		return err
	}

	e.Log.Infof("Token valid for trust domain: %s\n", t.Member.TrustDomain)

	return nil
}

func (e *Endpoints) handleError(w http.ResponseWriter, err error) {
	errMsg := fmt.Sprintf("failed processing request: %v", err)
	e.Log.Errorf(errMsg)
	w.WriteHeader(500)
	w.Write([]byte(errMsg))
}
