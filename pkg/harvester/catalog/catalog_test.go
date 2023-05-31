package catalog

import (
	"fmt"
	"os"
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"github.com/HewlettPackard/galadriel/test/certtest"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/require"
)

var hclConfigTemplate = `
providers {
    BundleSigner "disk" {
		ca_cert_path = "%s"
        ca_private_key_path = "%s"
	}
    BundleVerifier "disk" {
		trust_bundle_path = "%s"
	}
BundleVerifier "noop" {
	}
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
	require.NotNil(t, pc.BundleSigner)
	require.NotNil(t, pc.BundleVerifiers)
}

func TestLoadFromProvidersConfig(t *testing.T) {
	tempDir, cleanup := setupTest(t)
	defer cleanup()

	hclConfig := fmt.Sprintf(hclConfigTemplate, tempDir+"/root-ca.crt", tempDir+"/root-ca.key", tempDir+"/root-ca.crt")

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
	require.NotNil(t, cat.GetBundleSigner())
	require.NotNil(t, cat.GetBundleVerifiers())

	_, ok := cat.GetBundleSigner().(*integrity.DiskSigner)
	require.True(t, ok)
	_, ok = cat.GetBundleVerifiers()[0].(*integrity.DiskVerifier)
	require.True(t, ok)
	_, ok = cat.GetBundleVerifiers()[1].(*integrity.NoOpVerifier)
	require.True(t, ok)
}

func setupTest(t *testing.T) (string, func()) {
	tempDir := certtest.CreateTestCACertificates(t, clk)
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}
