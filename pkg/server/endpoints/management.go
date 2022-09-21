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

func (e *EndpointHandler) createMemberHandler(ctx context.Context) {
	http.HandleFunc("/createMember", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			e.Log.Errorf("failed reading body: %v", err)
			w.WriteHeader(500)
			return
		}
		memberReq := &common.Member{}
		err = json.Unmarshal(body, &memberReq)
		if err != nil {
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
		memberReq.Tokens = append(memberReq.Tokens, common.AccessToken{Token: token})
		m, err := e.DataStore.CreateMember(ctx, memberReq)
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
	})
}

func (e *EndpointHandler) createRelationshipHandler(ctx context.Context) {
	http.HandleFunc("/createRelationship", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			e.Log.Errorf("failed reading body: %v", err)
			w.WriteHeader(500)
			return
		}
		relationshipReq := &common.Relationship{}
		err = json.Unmarshal(body, &relationshipReq)
		if err != nil {
			e.Log.Errorf("failed unmarshalling body: %v", err)
			w.WriteHeader(500)
			return
		}

		rel, err := e.DataStore.CreateRelationship(ctx, relationshipReq)
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
	})
}

func (e *EndpointHandler) generateTokenHandler(ctx context.Context) {
	http.HandleFunc("/createToken", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			e.Log.Errorf("failed parsing body: %v", err)
			w.WriteHeader(500)
			return
		}
		m := &common.Member{}
		err = json.Unmarshal(body, m)
		if err != nil {
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

		at, err := e.DataStore.CreateAccessToken(
			ctx, &common.AccessToken{Token: token, Expiry: time.Now()}, m.ID,
		)
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
	})
}
