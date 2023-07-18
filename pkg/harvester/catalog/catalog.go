package catalog

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/jmhodges/clock"
)

// Catalog is a collection of provider interfaces.
type Catalog interface {
	GetBundleSigner() []integrity.Signer
	GetBundleVerifiers() []integrity.Verifier
}

// ProvidersRepository is the implementation of the Catalog interface.
type ProvidersRepository struct {
	bundleSigner    integrity.Signer
	bundleVerifiers []integrity.Verifier

	clock clock.Clock
}

// ProvidersConfig holds the HCL configuration for the providers.
type ProvidersConfig struct {
	BundleSigner    *providerConfig   `hcl:"BundleSigner,block"`
	BundleVerifiers []*providerConfig `hcl:"BundleVerifier,block"`
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
	if config.BundleSigner == nil {
		return fmt.Errorf("BundleSigner configuration is required")
	}
	if len(config.BundleVerifiers) == 0 {
		return fmt.Errorf("BundleVerifiers configuration is required")
	}

	var err error
	c.bundleSigner, err = loadBundleSigner(config.BundleSigner, c.clock)
	if err != nil {
		return fmt.Errorf("error loading BundleSigner: %w", err)
	}

	for _, bv := range config.BundleVerifiers {
		bundleVerifier, err := loadBundleVerifier(bv, c.clock)
		if err != nil {
			return fmt.Errorf("error loading BundlerVerifier: %w", err)
		}

		c.bundleVerifiers = append(c.bundleVerifiers, bundleVerifier)
	}

	return nil
}

func (c *ProvidersRepository) GetBundleSigner() integrity.Signer {
	return c.bundleSigner
}

func (c *ProvidersRepository) GetBundleVerifiers() []integrity.Verifier {
	return c.bundleVerifiers
}

func loadBundleSigner(config *providerConfig, clk clock.Clock) (integrity.Signer, error) {
	switch config.Name {
	case "disk":
		c, err := decodeDiskBundleSignerConfig(config)
		if err != nil {
			return nil, fmt.Errorf("error creating disk bundle verifier: %w", err)
		}
		c.Clock = clk

		verifier := integrity.NewDiskSigner()
		if err := verifier.Configure(c); err != nil {
			return nil, fmt.Errorf("error configuring disk bundle verifier: %w", err)
		}

		return verifier, nil
	case "noop":
		return integrity.NewNoOpSigner(), nil
	}

	return nil, fmt.Errorf("unknown bundle signer provider: %s", config.Name)
}

func decodeDiskBundleSignerConfig(config *providerConfig) (*integrity.DiskSignerConfig, error) {
	var dsConfig integrity.DiskSignerConfig
	if err := gohcl.DecodeBody(config.Options, nil, &dsConfig); err != nil {
		return nil, err
	}

	return &dsConfig, nil
}

func loadBundleVerifier(config *providerConfig, clk clock.Clock) (integrity.Verifier, error) {
	switch config.Name {
	case "disk":
		c, err := decodeDiskBundleVerifierConfig(config)
		if err != nil {
			return nil, fmt.Errorf("error creating disk bundle verifier: %w", err)
		}

		c.Clock = clk
		verifier := integrity.NewDiskVerifier()
		if err := verifier.Configure(c); err != nil {
			return nil, fmt.Errorf("error configuring disk bundle verifier: %w", err)
		}

		return verifier, nil
	case "noop":
		return integrity.NewNoOpVerifier(), nil
	}

	return nil, fmt.Errorf("unknown bundle signer provider: %s", config.Name)
}

func decodeDiskBundleVerifierConfig(config *providerConfig) (*integrity.DiskVerifierConfig, error) {
	var dsConfig integrity.DiskVerifierConfig
	if err := gohcl.DecodeBody(config.Options, nil, &dsConfig); err != nil {
		return nil, err
	}

	return &dsConfig, nil
}
