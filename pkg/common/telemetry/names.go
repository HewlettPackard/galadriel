package telemetry

// package name
const (
	Harvester = "harvester"
	Server    = "server"
)

// entity
const (
	TrustBundle            = "trust_bundle"
	PackageName            = "package_name"
	FederationRelationship = "federation_relationship"
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
	MetricsServer       = "metrics_server"
	HarvesterController = "harvester_controller"
	LocalSpireServer    = "local_spire_server"

	GaladrielServer = "galadriel_server"
	HTTPApi         = "http_api"
)
