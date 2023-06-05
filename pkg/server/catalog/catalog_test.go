package catalog

import (
	"fmt"
	"os"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/x509ca/disk"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

var hclConfigTemplate = `
providers {
    Datastore "sqlite3" {
		connection_string = "%s"
	}
	X509CA "disk" {
		key_file_path = "%s"
		cert_file_path = "%s"
	}
    KeyManager "memory" {}
}
`
var clk = clock.NewFake()

type providers struct {
	Block *providersBlock `hcl:"providers,block"`
}

type providersBlock struct {
	Body hcl.Body `hcl:",remain"`
}

func TestProvidersConfigsFromHCLBody(t *testing.T) {
	hclBody, diagErr := hclsyntax.ParseConfig([]byte(hclConfigTemplate), "", hcl.Pos{Line: 1, Column: 1})
	require.False(t, diagErr.HasErrors())

	var providers providers
	diagErr = gohcl.DecodeBody(hclBody.Body, nil, &providers)
	require.False(t, diagErr.HasErrors())

	pc, err := ProvidersConfigsFromHCLBody(providers.Block.Body)
	require.NoError(t, err)
	require.NotNil(t, pc)
	require.NotNil(t, pc.X509CA)
}

func TestLoadFromProvidersConfig(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	hclConfig := fmt.Sprintf(hclConfigTemplate, ":memory:", tempDir+"/root-ca.key", tempDir+"/root-ca.crt")

	hclBody, diagErr := hclsyntax.ParseConfig([]byte(hclConfig), "", hcl.Pos{Line: 1, Column: 1})
	require.False(t, diagErr.HasErrors())

	var providers providers
	diagErr = gohcl.DecodeBody(hclBody.Body, nil, &providers)
	require.False(t, diagErr.HasErrors())

	pc, err := ProvidersConfigsFromHCLBody(providers.Block.Body)
	require.NoError(t, err)
	require.NotNil(t, pc)

	cat := New()
	err = cat.LoadFromProvidersConfig(pc)
	require.NoError(t, err)
	require.NotNil(t, cat.GetDatastore())
	require.NotNil(t, cat.GetKeyManager())
	require.NotNil(t, cat.GetX509CA())

	_, ok := cat.GetX509CA().(*disk.X509CA)
	require.True(t, ok)
}

func setupTest(t *testing.T) (string, func()) {
	tempDir := certtest.CreateTestCACertificates(t, clk)
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}
