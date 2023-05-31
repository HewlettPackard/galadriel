package catalog

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/harvester/integrity"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
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

type diskBundleSignerConfig struct {
	CACertPath       string `hcl:"ca_cert_path"`
	CAPrivateKeyPath string `hcl:"ca_private_key_path"`
}

type diskBundleVerifierConfig struct {
	TrustBundlePath string `hcl:"trust_bundle_path"`
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
	c.bundleSigner, err = loadBundleSigner(config.BundleSigner)
	if err != nil {
		return fmt.Errorf("error loading BundleSigner: %w", err)
	}

	for _, vc := range config.BundleVerifiers {
		bundleVerifier, err := loadBundleVerifier(vc)
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

func loadBundleSigner(config *providerConfig) (integrity.Signer, error) {
	switch config.Name {
	case "disk":
		c, err := decodeDiskBundleSignerConfig(config)
		if err != nil {
			return nil, fmt.Errorf("error creating disk bundle verifier: %w", err)
		}
		verifier, err := integrity.NewDiskSigner(&integrity.DiskSignerConfig{
			CACertPath:       c.CACertPath,
			CAPrivateKeyPath: c.CAPrivateKeyPath,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating disk bundle signer: %w", err)
		}
		return verifier, nil
	case "noop":
		return integrity.NewNoOpSigner(), nil
	}

	return nil, fmt.Errorf("unknown bundle signer provider: %s", config.Name)
}

func decodeDiskBundleSignerConfig(config *providerConfig) (*diskBundleSignerConfig, error) {
	var dsConfig diskBundleSignerConfig
	if err := gohcl.DecodeBody(config.Options, nil, &dsConfig); err != nil {
		return nil, err
	}
	return &dsConfig, nil
}

func loadBundleVerifier(config *providerConfig) (integrity.Verifier, error) {
	switch config.Name {
	case "disk":
		c, err := decodeDiskBundleVerifierConfig(config)
		if err != nil {
			return nil, fmt.Errorf("error creating disk bundle verifier: %w", err)
		}
		verifier, err := integrity.NewDiskVerifier(&integrity.DiskVerifierConfig{
			TrustBundlePath: c.TrustBundlePath,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating disk bundle verifier: %w", err)
		}
		return verifier, nil
	case "noop":
		return integrity.NewNoOpVerifier(), nil
	}

	return nil, fmt.Errorf("unknown bundle signer provider: %s", config.Name)
}

func decodeDiskBundleVerifierConfig(config *providerConfig) (*diskBundleVerifierConfig, error) {
	var dsConfig diskBundleVerifierConfig
	if err := gohcl.DecodeBody(config.Options, nil, &dsConfig); err != nil {
		return nil, err
	}
	return &dsConfig, nil
}
