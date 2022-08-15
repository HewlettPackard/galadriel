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
)

// action
const (
	Add    = "add"
	Get    = "get"
	Remove = "remove"
	List   = "list"
	Create = "create"
)

// component
const (
	MetricsServer       = "metrics_server"
	HarvesterController = "harvester_controller"

	GaladrielServer = "galadriel_server"
	HTTPApi         = "http_api"
)
