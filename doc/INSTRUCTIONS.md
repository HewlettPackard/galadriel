# Instructions
Below is a list of instructions for running the PoC application.

## Requirements
In order to run Galadriel you should have:
- [Go Lang](https://go.dev/dl/) installed at version `1.19.x`
- A running [SPIRE](https://spiffe.io/docs/latest/deploying/install-server/) server

## Running the PoC locally
In order to run the PoC locally, clone the repository:
```bash
git clone https://github.com/HewlettPackard/galadriel.git && cd galadriel
```

After cloning the repository you will be able to build the application:
```bash
make build
```

With the built application you can use the binaries in the `bin` directory to run the Galadriel Server and Harvester:

## Configuring the Galadriel Server and Harvester
Before continuing make sure you have configured the Galadriel [Server](./USAGE.md#galadriel-server-configuration-file) and [Harvester](./USAGE.md#galadriel-harvester-configuration-file) with the appropriate configuration for your environment.

## Galadriel Server
To start the Galadriel Server you can use:
```bash
bin/galadriel-server run
```

You should see something like this in your terminal, indicating that the Galadriel Server is now `running`
```bash
INFO[0000] Starting TCP Server on 127.0.0.1:8085         subsystem_name=endpoints
INFO[0000] Starting UDS Server on /tmp/galadriel-server/api.sock  subsystem_name=endpoints
```

With the Galadriel Server running you will need to register a new Galadriel Harvester `Member`:
```bash
bin/galadriel-server create member -t <your SPIRE Trust Domain>
```

After registering the `Member` you will need to generate a new token to onboard the Galadriel Harvester that will manage the SPIRE Server:
```bash
ACCESS_TOKEN=$(bin/galadriel-server generate token -t <your SPIRE Trust Domain> | cut -d ' ' -f 3)
```
## Galadriel Harvester
To start the Galadriel Harvester you can execute the following command, using the `Access Token` generated from the Galadriel Server:

```bash
bin/galadriel-harvester run -t $ACCESS_TOKEN
```

This will result in the following output:
```bash
INFO[0000] Starting Harvester                            subsystem_name=harvester
INFO[0000] Connected to Galadriel Server                 subsystem_name=galadriel_server_client
INFO[0000] Starting harvester controller                 subsystem_name=harvester_controller
```