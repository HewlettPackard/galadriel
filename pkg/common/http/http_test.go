package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/entity"
	"github.com/google/uuid"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/stretchr/testify/assert"
)

func FakeTrustDomain() (*entity.TrustDomain, error) {

	tdName, err := spiffeid.TrustDomainFromString("spiffe://teste")
	if err != nil {
		return nil, err
	}

	return &entity.TrustDomain{
		ID:   uuid.NullUUID{},
		Name: tdName,
	}, nil
}

func FakeRequest() (http.Request, error) {

	td, err := FakeTrustDomain()
	if err != nil {
		return http.Request{}, err
	}

	b, err := json.Marshal(td)
	if err != nil {
		return http.Request{}, err
	}

	reader := bytes.NewReader(b)
	closerReader := io.NopCloser(reader)

	return http.Request{
		Body: closerReader,
	}, nil

}

func TestFromJSBody(t *testing.T) {

	t.Run("Should parse properly", func(t *testing.T) {
		request, err := FakeRequest()
		assert.Nil(t, err)

		td := &entity.TrustDomain{ID: uuid.NullUUID{}}
		td, err = FromJSBody(&request, td)
		assert.Nil(t, err)

		fakeTD, err := FakeTrustDomain()
		assert.Nil(t, err)

		assert.Equal(t, *fakeTD, *td)
	})
}
