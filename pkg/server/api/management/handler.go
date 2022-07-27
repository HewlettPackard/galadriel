package management

import (
	"errors"

	"github.com/labstack/echo/v4"
)

type MyDumbServer struct {
}

// (GET /spireServers)
func (server MyDumbServer) GetSpireServers(ctx echo.Context, params GetSpireServersParams) error {
	return errors.New("Ah, não, deu erro")
}

// (POST /spireServers)
func (server MyDumbServer) CreateSpireServer(ctx echo.Context) error {
	return nil
}

// (DELETE /spireServers/{spireServerId})
func (server MyDumbServer) DeleteSpireServer(ctx echo.Context, spireServerId int64) error {
	return nil
}

// (PUT /spireServers/{spireServerId})
func (server MyDumbServer) UpdateSpireServer(ctx echo.Context, spireServerId int64) error {
	return nil
}

// (PUT /trustBundles/{trustBundleId})
func (server MyDumbServer) UpdateTrustBundle(ctx echo.Context, trustBundleId int64) error {
	return nil
}
