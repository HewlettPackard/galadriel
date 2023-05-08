package catalog

import (
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager/memory"

	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca/disk"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// Catalog is a collection of provider interfaces.
type Catalog interface {
	GetX509CA() x509ca.X509CA
	GetKeyManager() keymanager.KeyManager
}

// ProvidersRepository is the implementation of the Catalog interface.
type ProvidersRepository struct {
	x509ca     x509ca.X509CA
	keyManager keymanager.KeyManager
}

// ProvidersConfig holds the HCL configuration for the providers.
type ProvidersConfig struct {
	X509CA     *providerConfig `hcl:"X509CA,block"`
	KeyManager *providerConfig `hcl:"KeyManager,block"`
}

// providerConfig holds the HCL configuration options for a single provider.
type providerConfig struct {
	Name    string   `hcl:",label"`
	Options hcl.Body `hcl:",remain"`
}

// New creates a new ProvidersRepository.
// It is the responsibility of the caller to load the catalog with providers using LoadFromProvidersConfig.
func New() *ProvidersRepository {
	return &ProvidersRepository{}
}

// ProvidersConfigsFromHCLBody parses the HCL body and returns the providers configuration.
func ProvidersConfigsFromHCLBody(body hcl.Body) (*ProvidersConfig, error) {
	var providersConfig ProvidersConfig

	if err := gohcl.DecodeBody(body, nil, &providersConfig); err != nil {
		return nil, fmt.Errorf("error decoding providersConfig config: %w", err)
	}

	return &providersConfig, nil
}

// LoadFromProvidersConfig loads the catalog from HCL configuration.
func (c *ProvidersRepository) LoadFromProvidersConfig(config *ProvidersConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is required")
	}
	if config.X509CA == nil {
		return fmt.Errorf("X509CA configuration is required")
	}
	if config.KeyManager == nil {
		return fmt.Errorf("KeyManager configuration is required")
	}

	var err error
	c.x509ca, err = loadX509CA(config.X509CA)
	if err != nil {
		return fmt.Errorf("error loading X509CA: %w", err)
	}

	c.keyManager, err = loadKeyManager(config.KeyManager)
	if err != nil {
		return fmt.Errorf("error loading KeyManager: %w", err)
	}

	return nil
}

func (c *ProvidersRepository) GetX509CA() x509ca.X509CA {
	return c.x509ca
}

func (c *ProvidersRepository) GetKeyManager() keymanager.KeyManager {
	return c.keyManager
}

func loadX509CA(c *providerConfig) (x509ca.X509CA, error) {
	switch c.Name {
	case "disk":
		x509CA, err := makeDiskX509CA(c)
		if err != nil {
			return nil, fmt.Errorf("error creating disk X509CA: %w", err)
		}
		return x509CA, nil
	}

	return nil, fmt.Errorf("unknown X509CA provider: %s", c.Name)
}

func loadKeyManager(c *providerConfig) (keymanager.KeyManager, error) {
	switch c.Name {
	case "memory":
		km := memory.New(nil)
		return km, nil
	}

	return nil, fmt.Errorf("unknown KeyManager provider: %s", c.Name)
}

func makeDiskX509CA(config *providerConfig) (*disk.X509CA, error) {
	var diskX509CAConfig disk.Config
	if err := gohcl.DecodeBody(config.Options, nil, &diskX509CAConfig); err != nil {
		return nil, err
	}

	ca, err := disk.New()
	if err != nil {
		return nil, err
	}
	if err := ca.Configure(&diskX509CAConfig); err != nil {
		return nil, err
	}
	return ca, nil
}
