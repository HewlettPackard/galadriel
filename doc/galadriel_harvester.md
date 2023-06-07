# Galadriel Harvester Configuration Reference

This document provides a reference for the Galadriel Harvester configuration file.

## Configuration File

The Galadriel Harvester configuration file contains several sections that allow you to customize the behavior of the
Harvester.

### `harvester`

| Option                            | Description                                                                                                        | Default                              |
|-----------------------------------|--------------------------------------------------------------------------------------------------------------------|--------------------------------------|
| `trust_domain`                    | Specifies the trust domain of the SPIRE Server instance that the Harvester runs alongside.                         |                                      |
| `harvester_socket_path`           | Specifies the path to the UNIX Domain Socket that the Galadriel Harvester will listen on.                          | `/tmp/galadriel-harvester/api.sock`  |
| `spire_socket_path`               | Specifies the path to the UNIX Domain Socket of the SPIRE Server that the Harvester will connect to.               | `/tmp/spire-server/private/api.sock` |
| `galadriel_server_address`        | Specifies the DNS name or IP address and port of the upstream Galadriel Server that the Harvester will connect to. |                                      |
| `server_trust_bundle_path`        | Path to the Galadriel Server CA bundle.                                                                            |                                      |
| `federated_bundles_poll_interval` | Configure how often the harvester will poll federated bundles from the Galadriel Server.                           | `2m`                                 |
| `spire_bundle_poll_interval`      | Configure how often the harvester will poll the bundle from SPIRE.                                                 | `1m`                                 |
| `log_level`                       | Sets the logging level [DEBUG                                                                                      | INFO                                 |WARN|ERROR]. | `INFO` |
| `data_dir`                        | Directory to store persistent data.                                                                                |                                      |

### `providers`

| Provider         | Description                                                                                            |
|------------------|--------------------------------------------------------------------------------------------------------|
| `BundleSigner`   | Enables the signing of bundles using a selected implementation. Can be `noop` or `disk`.               |
| `BundleVerifier` | Enables the verification of bundle signatures using selected implementations. Can be `noop` or `disk`. |

#### BundleSigner

| Option | Description                                                                                                                                                                              |
|--------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `noop` | No-op signing (doesn't sign bundles).                                                                                                                                                    |
| `disk` | Enables the signing of bundles using a disk-based key pair. The `ca_cert_path` is the path to the CA certificate file. The `ca_private_key_path` is the path to the CA private key file. |

#### Example:

```hcl
providers {
  BundleSigner "disk" {
    ca_cert_path = "conf/harvester/dummy_root_ca.crt"
    ca_private_key_path = "conf/harvester/dummy_root_ca.key"
  }
}
```

#### BundleVerifier

| Option | Description                                                                                                                                  |
|--------|----------------------------------------------------------------------------------------------------------------------------------------------|
| `noop` | If this verifier is enabled, all bundles will pass the verification process without actually validating the signatures.                      |
| `disk` | Enables the verification of bundle signatures using a disk-based trust bundle. The `trust_bundle_path` is the path to the trust bundle file. |

#### Example:

```hcl
providers {
  BundleVerifier "disk" {
    trust_bundle_path = "conf/harvester/dummy_root_ca.crt"
  }
}
```

## Galadriel Harvester CLI Reference

The Galadriel Harvester provides a command-line interface (CLI) for running the Harvester and managing relationships for
the trust domain managed by the SPIRE Server that this Harvester runs alongside.

To access the CLI, use the `galadriel-harvester` command:

```bash
./galadriel-harvester
```

### Available Commands

#### `run`

Run this command to start the Galadriel Harvester.

```bash
./galadriel-harvester run [flags]
```

| Flag              | Description                                  | Default                         |
|-------------------|----------------------------------------------|---------------------------------|
| `-c, --config`    | Path to the Galadriel Harvester config file. | `conf/harvester/harvester.conf` |
| `-t, --joinToken` | A join token generated by Galadriel Server.  |                                 |

#### `relationship`

The 'relationship' command allows you to manage

relationships within the trust domain managed by the SPIRE Server that this Harvester runs alongside.

```bash
./galadriel-harvester relationship [command]
```

Available Commands:

- `approve`: Approve a relationship.
- `deny`: Deny a relationship.
- `list`: List relationships for the trust domain managed by the SPIRE Server this Harvester runs alongside.

### Global Flags

| Flag           | Description                                 | Default                             |
|----------------|---------------------------------------------|-------------------------------------|
| `--socketPath` | Path to the Galadriel Harvester API socket. | `/tmp/galadriel-harvester/api.sock` |

## Sample Configuration File

Below is a sample configuration file for the Galadriel Harvester. This file includes examples of how to configure the
Harvester, providers, and other options.

```hcl
harvester {
  trust_domain = "example.org"
  harvester_socket_path = "/tmp/galadriel-harvester/api.sock"
  spire_socket_path = "/tmp/spire-server/private/api.sock"
  galadriel_server_address = "localhost:8085"
  server_trust_bundle_path = "./conf/harvester/dummy_root_ca.crt"
  federated_bundles_poll_interval = "10s"
  spire_bundle_poll_interval = "10s"
  log_level = "DEBUG"
  data_dir = "./.data"
}

providers {
  BundleSigner "disk" {
    ca_cert_path = "conf/harvester/dummy_root_ca.crt"
    ca_private_key_path = "conf/harvester/dummy_root_ca.key"
  }

  BundleVerifier "disk" {
    trust_bundle_path = "conf/harvester/dummy_root_ca.crt"
  }
}
```
