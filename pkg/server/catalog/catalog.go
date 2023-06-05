package catalog

import (
	"fmt"

	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca"
	"github.com/HewlettPackard/galadriel/pkg/common/x509ca/disk"
	"github.com/HewlettPackard/galadriel/pkg/server/db"
	"github.com/HewlettPackard/galadriel/pkg/server/db/postgres"
	"github.com/HewlettPackard/galadriel/pkg/server/db/sqlite"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
)

// Catalog is a collection of provider interfaces.
type Catalog interface {
	GetDatastore() db.Datastore
	GetX509CA() x509ca.X509CA
	GetKeyManager() keymanager.KeyManager
}

// ProvidersRepository is the implementation of the Catalog interface.
type ProvidersRepository struct {
	datastore  db.Datastore
	x509ca     x509ca.X509CA
	keyManager keymanager.KeyManager
}

// ProvidersConfig holds the HCL configuration for the providers.
type ProvidersConfig struct {
	Datastore  *providerConfig `hcl:"Datastore,block"`
	X509CA     *providerConfig `hcl:"X509CA,block"`
	KeyManager *providerConfig `hcl:"KeyManager,block"`
}

// providerConfig holds the HCL configuration options for a single provider.
type providerConfig struct {
	Name    string   `hcl:",label"`
	Options hcl.Body `hcl:",remain"`
}

type datastoreConfig struct {
	ConnectionString string `hcl:"connection_string"`
}

type diskKeyManagerConfig struct {
	KeysFilePath string `hcl:"keys_file_path"`
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
	if config.Datastore == nil {
		return fmt.Errorf("datastore configuration is required")
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
	c.datastore, err = loadDatastore(config.Datastore)
	if err != nil {
		return fmt.Errorf("error loading datastore: %w", err)
	}

	return nil
}

func (c *ProvidersRepository) GetDatastore() db.Datastore {
	return c.datastore
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
		km := keymanager.NewMemoryKeyManager(nil)
		return km, nil
	case "disk":
		kmConfig, err := decodeDiskKeyManagerConfig(c)
		if err != nil {
			return nil, fmt.Errorf("error decoding disk KeyManager config: %w", err)
		}
		km, err := keymanager.NewDiskKeyManager(nil, kmConfig.KeysFilePath)
		if err != nil {
			return nil, fmt.Errorf("error creating disk KeyManager: %w", err)
		}
		return km, nil
	}

	return nil, fmt.Errorf("unknown KeyManager provider: %s", c.Name)
}

func loadDatastore(config *providerConfig) (db.Datastore, error) {
	c, err := decodeDatastoreConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error decoding datastore config: %w", err)
	}
	switch config.Name {
	case "postgres":
		ds, err := postgres.NewDatastore(c.ConnectionString)
		if err != nil {
			return nil, fmt.Errorf("error creating postgres datastore: %w", err)
		}
		return ds, nil
	case "sqlite3":
		ds, err := sqlite.NewDatastore(c.ConnectionString)
		if err != nil {
			return nil, fmt.Errorf("error creating sqlite datastore: %w", err)
		}

		return ds, nil

	}

	return nil, fmt.Errorf("unknown datastore provider: %s", config.Name)
}

func decodeDatastoreConfig(config *providerConfig) (*datastoreConfig, error) {
	var dsConfig datastoreConfig
	if err := gohcl.DecodeBody(config.Options, nil, &dsConfig); err != nil {
		return nil, err
	}
	return &dsConfig, nil
}

func decodeDiskKeyManagerConfig(config *providerConfig) (*diskKeyManagerConfig, error) {
	var dsConfig diskKeyManagerConfig
	if err := gohcl.DecodeBody(config.Options, nil, &dsConfig); err != nil {
		return nil, err
	}
	return &dsConfig, nil
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
