package telemetry

const (
	// ElapsedTime tags some duration of time.
	ElapsedTime = "elapsed_time"

	// Status tags status of call (OK, or some error), or status of some process
	Status = "status"

	// Catalog functionality related to plugin catalog
	Catalog = "catalog"

	// Endpoints related to API endpoints
	Endpoints = "endpoints"

	// HarvesterController functionality related to Harvester controller
	HarvesterController = "harvester_controller"

	// GaladrielServer tags the Galadriel server module
	GaladrielServer = "galadriel_server"

	// SubsystemName declares a field for some subsystem name (an API, module...)
	SubsystemName = "subsystem_name"

	// GaladrielServerClient functionality related to Galadriel server client
	GaladrielServerClient = "galadriel_server_client"

	// Telemetry tags a telemetry module
	Telemetry = "telemetry"

	// Harvester tags the Harvester module
	Harvester = "harvester"
)
