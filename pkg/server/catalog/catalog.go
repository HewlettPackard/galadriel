package catalog

import (
	"fmt"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca/disk"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// Catalog is a collection of provider interfaces.
type Catalog interface {
	GetX509CA() x509ca.X509CA
}

// ProvidersRepository is the implementation of the Catalog interface.
type ProvidersRepository struct {
	x509ca x509ca.X509CA
}

// ProvidersConfig holds the HCL configuration for the providers.
type ProvidersConfig struct {
	X509CA *providerConfig `hcl:"x509ca,block"`
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

	switch config.X509CA.Name {
	case "disk":
		x509CA, err := makeDiskX509CA(config.X509CA)
		if err != nil {
			return fmt.Errorf("error creating disk X509CA: %w", err)
		}
		c.x509ca = x509CA
	default:
		return fmt.Errorf("unknown X509CA provider: %s", config.X509CA.Name)
	}

	return nil
}

func (c *ProvidersRepository) GetX509CA() x509ca.X509CA {
	return c.x509ca
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
