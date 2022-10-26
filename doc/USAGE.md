# Galadriel Server CLI
The Galadriel Server CLI contains the functionality to:
### `galadriel-server create member`
| Flag      |  Type  | Description               |
|-----------|--------|---------------------------|
| `-t`      | string | SPIRE server trust domain |


### `galadriel-server create relationship`
| Flag      |  Type  | Description                 |
|:----------|:-------|:----------------------------|
| `-a`      | string | SPIRE Server trust domain A |
| `-b`      | string | SPIRE Server trust domain A |


### `galadriel-server generate token`
| Flag      |  Type  | Description               |
|-----------|--------|---------------------------|
| `-t`      | string | SPIRE server trust domain |


### `galadriel-server list`
| Command         | Description                                           |
|-----------------|-------------------------------------------------------|
| `members`       | List all members stored in the Galadriel Server       |
| `relationships` | List all relationships stored in the Galadriel Server |

# Galadriel Harvester CLI
The Galadriel Harvester CLI contains the functionality to run the Galadriel Harvester while attaching it to the Galadriel Server instance, based on the token used as a argument:

### `galadriel-harvester run`
| Flag      |  Type  | Description               |
|-----------|--------|---------------------------|
| `-t`      | string | SPIRE server trust domain |

# Galadriel Server Configuration File
You can find the Galadriel Server configuration file at `galadriel/conf/server`

| Configuration    |  Description                                     | Default                         |
|------------------|--------------------------------------------------|---------------------------------|
| `listen_address` | IP address or DNS name of the Galadriel server.  |  localhost                      |
| `listen_port`    | HTTP Port number of the Galadriel server.        |  8085                           |
| `socket_path`    | Path to bind the Galadriel Server API socket to. |  /tmp/galadriel-server/api.sock |

# Galadriel Harvester Configuration File
You can find the Galadriel Server configuration file at `galadriel/conf/harvester`

| Configuration               |  Description                                                      | Default                             |
|-----------------------------|-------------------------------------------------------------------|-------------------------------------|
| `spire_socket_path`         | SPIRE Server Socket of the instance to manage.                    |  /tmp/spire-server/private/api.sock |
| `server_address`            | Upstream Galadriel Server DNS name or IP address with port.       |  localhost:8085                     |
| `bundle_updates_interval`   | Sets how often to check for bundle rotation.                      |  5s                                 |