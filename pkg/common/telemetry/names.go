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

	// FederatedBundlesSynchronizer represents the Federated Bundles Synchronizer subsystem.
	FederatedBundlesSynchronizer = "federated_bundles_synchronizer"

	// GaladrielServer represents the Galadriel server subsystem.
	GaladrielServer = "galadriel_server"

	// Network represents a network name ("tcp", "udp").
	Network = "network"

	// SpireBundleSynchronizer represents the SPIRE Bundle Synchronizer subsystem.
	SpireBundleSynchronizer = "spire_bundle_synchronizer"

	// SubsystemName represents a field for some subsystem name, such as an API or module.
	SubsystemName = "subsystem_name"

	// TrustDomain tags the name of some trust domain
	TrustDomain = "trust_domain"
)
