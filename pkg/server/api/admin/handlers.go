package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
)

func (s *localServer) createMemberHandler(w http.ResponseWriter, r *http.Request) {
	logger := s.config.Logger
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("failed reading body: %v", err)
		w.WriteHeader(500)
		return
	}
	memberReq := &common.Member{}
	err = json.Unmarshal(body, memberReq)
	if err != nil {
		logger.Errorf("failed unmarshalling body: %v", err)
		w.WriteHeader(500)
		return
	}

	m, err := s.DataStore.CreateMember(context.TODO(), memberReq)
	if err != nil {
		logger.Errorf("failed creating member: %v", err)
		w.WriteHeader(500)
		return
	}

	memberBytes, err := json.Marshal(m)
	if err != nil {
		logger.Errorf("failed marshalling member: %v", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	_, err = w.Write(memberBytes)
	if err != nil {
		return
	}

}

func (s *localServer) createRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	logger := s.config.Logger

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("failed reading body: %v", err)
		s.handleError(w, err)
		return
	}
	relationshipReq := common.Relationship{}
	err = json.Unmarshal(body, &relationshipReq)
	if err != nil {
		logger.Errorf("failed unmarshalling body: %v", err)
		s.handleError(w, err)
		return
	}

	rel, err := s.DataStore.CreateRelationship(context.TODO(), &relationshipReq)
	if err != nil {
		logger.Errorf("failed creating relationship: %v", err)
		s.handleError(w, err)
		return
	}

	relBytes, err := json.Marshal(rel)
	if err != nil {
		logger.Errorf("failed marshalling relationship: %v", err)
		s.handleError(w, err)
		return
	}

	_, err = w.Write(relBytes)
	if err != nil {
		return
	}
}

func (s *localServer) generateTokenHandler(w http.ResponseWriter, r *http.Request) {
	logger := s.config.Logger

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("failed parsing body: %v", err)
		s.handleError(w, err)
		return
	}
	var member common.Member
	err = json.Unmarshal(body, &member)
	if err != nil {
		logger.Errorf("failed unmarshalling body: %v", err)
		s.handleError(w, err)
		return
	}

	token, err := util.GenerateToken()
	if err != nil {
		logger.Errorf("failed generating token: %v", err)
		s.handleError(w, err)
		return
	}

	at, err := s.DataStore.GenerateAccessToken(context.TODO(), &common.AccessToken{
		Token: token, Expiry: time.Now(),
	}, member.TrustDomain.String())
	if err != nil {
		logger.Errorf("failed creating access token: %v", err)
		s.handleError(w, err)
		return
	}
	tokenBytes, err := json.Marshal(at)
	if err != nil {
		logger.Errorf("failed marshalling token: %v", err)
		s.handleError(w, err)
		return
	}

	_, err = w.Write(tokenBytes)
	if err != nil {
		logger.Errorf("failed to return token: %v", err)
		s.handleError(w, err)
		return
	}
}

func (s *localServer) handleError(w http.ResponseWriter, err error) {
	errMsg := fmt.Sprintf("failed processing request: %v", err)
	s.Logger.Errorf(errMsg)
	w.WriteHeader(500)
	_, _ = w.Write([]byte(errMsg))
}
