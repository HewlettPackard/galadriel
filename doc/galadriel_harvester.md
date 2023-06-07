# Galadriel Harvester Configuration and CLI Reference

This document provides a comprehensive reference for both the Galadriel Harvester configuration file and its
command-line interface (CLI). It details each section of the configuration file and explains various CLI commands to
assist with the
harvester's setup, customization, and management.

## Configuration File

The Galadriel Harvester configuration file contains various sections that let you customize the Harvester's behaviour,
enhancing your control over its functions and operations.

### `harvester`

This section lists the options available for configuring the main behavior of the Galadriel Harvester.

| Option                            | Description                                                                                                        | Default                              |
|-----------------------------------|--------------------------------------------------------------------------------------------------------------------|--------------------------------------|
| `trust_domain`                    | Specifies the trust domain of the SPIRE Server instance that the Harvester runs alongside.                         |                                      |
| `harvester_socket_path`           | Specifies the path to the UNIX Domain Socket that the Galadriel Harvester will listen on.                          | `/tmp/galadriel-harvester/api.sock`  |
| `spire_socket_path`               | Specifies the path to the UNIX Domain Socket of the SPIRE Server that the Harvester will connect to.               | `/tmp/spire-server/private/api.sock` |
| `galadriel_server_address`        | Specifies the DNS name or IP address and port of the upstream Galadriel Server that the Harvester will connect to. |                                      |
| `server_trust_bundle_path`        | Path to the Galadriel Server CA bundle that will be used to verify the Server's certificate.                       |                                      |
| `federated_bundles_poll_interval` | Configure how often the harvester will poll federated bundles from the Galadriel Server.                           | `2m`                                 |
| `spire_bundle_poll_interval`      | Configure how often the harvester will poll the bundle from SPIRE.                                                 | `1m`                                 |
| `log_level`                       | Sets the logging level. Options are `DEBUG`, `WARN`, `INFO`, `ERROR`                                               | `INFO`                               |
| `data_dir`                        | Directory to store persistent data.                                                                                |                                      |

### `providers`

This section describes the configuration options for the `BundleSigner` and `BundleVerifier` providers in the Galadriel
Harvester.

| Provider         | Description                                                                                            |
|------------------|--------------------------------------------------------------------------------------------------------|
| `BundleSigner`   | Enables the signing of bundles using a selected implementation. Can be `noop` or `disk`.               |
| `BundleVerifier` | Enables the verification of bundle signatures using selected implementations. Can be `noop` or `disk`. |

#### BundleSigner

This subsection illustrates options available for the `BundleSigner`.

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

This subsection explains the `BundleVerifier` options.

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

The Galadriel Harvester provides a command-line interface (CLI) for operating the Harvester and managing
relationships for the trust domain overseen by the corresponding SPIRE Server.

To access the CLI, you can use the `galadriel-harvester` command:

```bash
./galadriel-harvester
```

### Available Commands

This section describes the available commands and their usage.

#### `run`

This command starts the Galadriel Harvester.

```bash
./galadriel-harvester run [flags]
```

| Flag              | Description                                                                               | Default                         |
|-------------------|-------------------------------------------------------------------------------------------|---------------------------------|
| `-c, --config`    | Path to the Galadriel Harvester config file.                                              | `conf/harvester/harvester.conf` |
| `-t, --joinToken` | A join token generated by Galadriel Server used to introduce the Harvester to the Server. |                                 |

#### `relationship`

The 'relationship' command assists you in managing relationships within the trust domain regulated by the SPIRE Server
that the Harvester operates with.

```bash
./galadriel-harvester relationship [command]
```

#### Available subcommands:

- `approve` - Authorize participation in the Federation relationship.
- `deny` - Refuse participation in the Federation relationship.
- `list` - List all relationships for the trust domain managed by the SPIRE Server that the Harvester operates with.

##### `relationship approve`

The `approve` command is for approving a relationship for the trust domain of the SPIRE Server that the Harvester
operates with.

Syntax:

```bash
./galadriel-harvester relationship approve [flags]
```

Example Usage:

```bash
./galadriel-harvester relationship approve --relationshipID <relationshipID>
```

| Flag                          | Description                                            | Default |
|-------------------------------|--------------------------------------------------------|---------|
| `-r, --relationshipID string` | The specific Relationship ID that you wish to approve. |         |

##### `relationship deny`

The `deny` command enables you to reject a relationship within the trust domain managed by the SPIRE Server that the
Harvester operates with.

Syntax:

```bash
./galadriel-harvester relationship deny [flags]
```

Example Usage:

```bash
./galadriel-harvester relationship deny --relationshipID <relationshipID>
```

| Flag                          | Description                                         | Default |
|-------------------------------|-----------------------------------------------------|---------|
| `-r, --relationshipID string` | The specific Relationship ID that you wish to deny. |         |

##### `relationship list`

The `list` command allows you to view all relationships within the trust domain managed by the SPIRE Server where the
Harvester operates.

```bash
./galadriel-harvester relationship list [flags]
```

Example Usage:

```bash
./galadriel-harvester relationship list
```

### Global Flags

These flags can be used across all commands.

| Flag           | Description                                 | Default                             |
|----------------|---------------------------------------------|-------------------------------------|
| `--socketPath` | Path to the Galadriel Harvester API socket. | `/tmp/galadriel-harvester/api.sock` |

## Sample Configuration File

Provided below is a sample configuration file for the Galadriel Harvester. It demonstrates how to configure the
Harvester, providers, and other available options.

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
