package telemetry

// entity
const (
	TrustBundle = "trust_bundle"
)

// action
const (
	Add    = "add"
	Remove = "remove"
	List   = "list"
	Create = "create"
)

// component
const (
	SpireServer     = "spire_server"
	Harvester       = "harvester"
	GaladrielServer = "galadriel_server"
)

// telemetry.StartCall(m, telemetry.Datastore, telemetry.Bundle, telemetry.Create)
