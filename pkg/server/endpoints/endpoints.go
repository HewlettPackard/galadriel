package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/util"
	"github.com/HewlettPackard/galadriel/pkg/server/datastore"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"time"
)

// Server manages the UDS and TCP endpoints lifecycle
type Server interface {
	// ListenAndServe starts all endpoint servers and blocks until the context
	// is canceled or any of the endpoints fails to run.
	ListenAndServe(ctx context.Context) error
}

type Endpoints struct {
	TCPAddress *net.TCPAddr
	LocalAddr  net.Addr
	DataStore  datastore.DataStore
	Log        logrus.FieldLogger
}

func New(c *Config) (*Endpoints, error) {
	if err := util.PrepareLocalAddr(c.LocalAddress); err != nil {
		return nil, err
	}
	return &Endpoints{
		TCPAddress: c.TCPAddress,
		LocalAddr:  c.LocalAddress,
		DataStore:  c.Catalog.GetDataStore(),
		Log:        c.Log,
	}, nil
}

func (e *Endpoints) ListenAndServe(ctx context.Context) error {
	tasks := []func(context.Context) error{
		e.runTCPServer,
		e.runUDSServer,
	}

	err := util.RunTasks(ctx, tasks)
	if err != nil {
		return err
	}

	return nil
}

func (e *Endpoints) runTCPServer(ctx context.Context) error {
	server := echo.New()

	e.Log.Info("Starting TCP Server")
	errChan := make(chan error)
	go func() {
		errChan <- server.Start(e.TCPAddress.String())
	}()

	var err error
	select {
	case err = <-errChan:
		e.Log.WithError(err).Error("TCP Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Log.Info("Stopping TCP Server")
		server.Close()
		<-errChan
		e.Log.Info("TCP Server stopped")
		return nil
	}
}

func (e *Endpoints) runUDSServer(ctx context.Context) error {
	server := &http.Server{}

	l, err := net.Listen(e.LocalAddr.Network(), e.LocalAddr.String())
	if err != nil {
		return fmt.Errorf("error listening on uds: %w", err)
	}
	defer l.Close()

	e.addHandlers(ctx)

	e.Log.Info("Starting UDS Server")
	errChan := make(chan error)
	go func() {
		errChan <- server.Serve(l)
	}()

	select {
	case err = <-errChan:
		e.Log.WithError(err).Error("Local Server stopped prematurely")
		return err
	case <-ctx.Done():
		e.Log.Info("Stopping UDS Server")
		server.Close()
		<-errChan
		e.Log.Info("UDS Server stopped")
		return nil
	}
}

func (e *Endpoints) addHandlers(ctx context.Context) {
	e.createMemberHandler(ctx)
	e.createRelationshipHandler(ctx)
	e.generateTokenHandler(ctx)
}

func (e *Endpoints) createMemberHandler(ctx context.Context) {
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

func (e *Endpoints) createRelationshipHandler(ctx context.Context) {
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

func (e *Endpoints) generateTokenHandler(ctx context.Context) {
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
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
