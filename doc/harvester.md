# Galadriel Harvester Configuration Reference

## Configuration file

| Configuration | Description | Default | Required
| -- | -- | -- | --
| `spire_socket_path` | Path to the SPIRE Server UDS of the instance to manage | `/tmp/spire-server/private/api.sock` |
| `server_address` | Upstream Galadriel Server DNS name or IP address with port. E.g `localhost:8080`, `my-upstream-server.com:4556`, `192.168.1.125:4000` | | Yes
| `log_level` | Logging level. One of: `DEBUG`, `INFO`, `WARN`, `ERROR` | `INFO` |

## Command line arguments

### `harvester run`

| Subcommand | Description | Default | Required
| -- | -- | -- | --
| `-config` | Path to the Harvester config file | `conf/harvester/harvester.conf` |
