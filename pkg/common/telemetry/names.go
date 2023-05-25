package telemetry

// package name
const (
	Harvester = "harvester"
	Server    = "server"
)

// entity
const (
	TrustBundle = "trust_bundle"
	PackageName = "package_name"
	Federation  = "federation"
)

// action
const (
	Add     = "add"
	Get     = "get"
	Remove  = "remove"
	List    = "list"
	Create  = "create"
	Approve = "approve"
	Deny    = "deny"
)

const (
	// Address represents a network address.
	Address = "address"

	// BundleOpStatus represents a bundle operation status.
	BundleOpStatus = "bundle_op_status"

	// DiskX509CA represents a disk-based X509 CA.
	DiskX509CA = "disk_x509_ca"

	// Endpoints represents functionality related to agent/server endpoints.
	Endpoints = "endpoints"

	// FederadBundlesSyncer represents the Federated Bundles Syncer subsystem.
	FederadBundlesSyncer = "federated_bundles_syncer"

	// GaladrielServer represents the Galadriel server subsystem.
	GaladrielServer = "galadriel_server"

	// Network represents a network name ("tcp", "udp").
	Network = "network"

	// SpireBundleSyncer represents the SPIRE Bundle Syncer subsystem.
	SpireBundleSyncer = "spire_bundle_syncer"

	// SubsystemName represents a field for some subsystem name, such as an API or module.
	SubsystemName = "subsystem_name"

	// TrustDomain tags the name of some trust domain
	TrustDomain = "trust_domain"
)
