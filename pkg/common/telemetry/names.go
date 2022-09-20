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

// component
const (
	// Catalog functionality related to plugin catalog
	Catalog = "catalog"

	Endpoints = "endpoints"

	MetricsServer       = "metrics_server"
	HarvesterController = "harvester_controller"

	GaladrielServer = "galadriel_server"
	HTTPApi         = "http_api"

	// SubsystemName declares a field for some subsystem name (an API, module...)
	SubsystemName = "subsystem_name"

	GaladrielServerClient = "galadriel_server_client"
)
