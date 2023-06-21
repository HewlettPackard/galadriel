# Galadriel Server Configuration Reference

This document provides a comprehensive reference for both the Galadriel Server configuration file and its command-line
interface (CLI). It details each section of the configuration file and explains various CLI commands to assist with the
server's setup, customization, and management.

## Introduction to the Configuration File

The Galadriel Server configuration file is instrumental in tailoring the behavior of the Galadriel Server. It's divided
into multiple sections, primarily `server` and `providers`, with `providers` further containing `Datastore`, `X509CA`,
and `KeyManager`.

### Server Configuration (`server`)

This section facilitates the configuration of the server's fundamental characteristics. It includes properties such
as `listen_address`, `listen_port`, `socket_path`, and `log_level`. Below is the detailed description for each property
along with their default values:

| Property         | Description                                                                                                                 | Default                          |
|------------------|-----------------------------------------------------------------------------------------------------------------------------|----------------------------------|
| `listen_address` | Specifies the IP address or DNS name that the Galadriel server will bind to for accepting network connections.              | `0.0.0.0`                        |
| `listen_port`    | Specifies the HTTP port number that the Galadriel server will listen on for incoming connections.                           | `8085`                           |
| `socket_path`    | Specifies the path to the UNIX Domain Socket that the Galadriel Server API will bind to for communication on the same host. | `/tmp/galadriel-server/api.sock` |
| `log_level`      | Sets the logging level. Options are `DEBUG`, `INFO`, `WARN`, `ERROR`.                                                       | `INFO`                           |

#### Example:

```hcl
server {
  listen_address = "localhost"
  listen_port = "8085"
  socket_path = "/tmp/galadriel-server/api.sock"
  log_level = "DEBUG"
}
```

### Provider Configuration (`providers`)

The `providers` section allows you to configure the Datastore, X509CA, and KeyManager providers. Each provider is
detailed below:

| Provider     | Description                                                                  |
|--------------|------------------------------------------------------------------------------|
| `Datastore`  | Configures the datastore provider.                                           |
| `X509CA`     | Configures the X509CA provider for signing TLS X.509 certificates.           |
| `KeyManager` | Configures the KeyManager for providing private keys for signing JWT tokens. |

The following subsections provide detailed configurations for each provider:

#### Datastore Configuration

The Datastore section covers the configuration details for SQLite3 and PostgreSQL datastores:

| Option     | Description                                                                                  |
|------------|----------------------------------------------------------------------------------------------|
| `sqlite3`  | Uses SQLite3 as the datastore. The `connection_string` is the database connection string.    |
| `postgres` | Uses PostgreSQL as the datastore. The `connection_string` is the database connection string. |

#### Example:

```hcl
providers {
  Datastore "sqlite3" {
    connection_string = "./datastore.sqlite3"
  }
}
```

#### X509CA Configuration

The X509CA section provides configuration details for X.509 CA providers:

| Option             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `disk`             | Uses a CA (either ROOT or INTERMEDIATE) and private key loaded from disk to issue X.509 certificates.                                                                                                                                                                                                                                                                                                                                                            |
| `key_file_path`    | Path to the CA private key file in PEM format. This path can be relative or absolute.                                                                                                                                                                                                                                                                                                                                                                            |
| `cert_file_path`   | Path to the CA certificate file in PEM format. If Galadriel is using a self-signed CA, cert_file_path should specify the path to a single PEM encoded certificate representing the CA certificate. If not self-signed, cert_file_path should specify the path to a file that must contain one or more certificates necessary to establish a valid certificate chain up the root certificates defined in bundle_file_path. This path can be relative or absolute. |
| `bundle_file_path` | Required when the cert_file_path does not contain a self-signed CA certificate. This is the path to the file containing one or more root CAs. This path can be relative or absolute.                                                                                                                                                                                                                                                                             |

#### Example:

```hcl
providers {
  X509CA "disk" {
    key_file_path = "./conf/server/dummy_root_ca.key"
    cert_file_path = "./conf/server/dummy_root_ca.crt"
    bundle_file_path = "./conf/server/root_ca.crt"
  }
}
```

In the example above, the bundle_file_path is set, indicating that the certificate used in cert_file_path isn't
self-signed and requires a chain of trust to a root CA.

#### KeyManager Configuration

The KeyManager section discusses the configuration details for key managers:

| Option   | Description                                                                                                                                     |
|----------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| `memory` | A key manager for generating keys and signing certificates that stores keys in memory.                                                          |
| `disk`   | A key manager for generating keys that stores keys on disk. The `keys_file_path` is the path to the file where the key manager will store keys. |

#### Example:

```hcl
providers {
  KeyManager "disk" {
    keys_file_path = "./keys.json"
  }
}
```

Sure, here is the improved "Galadriel Server CLI Reference" section:

## Galadriel Server CLI Reference

The Galadriel server provides a command-line interface (CLI) for operating the server, managing federation
relationships, creating join tokens, and managing SPIFFE trust domains.

To access the CLI, use the `galadriel-server` command:

```bash
./galadriel-server
```

### CLI Commands

Below are the primary commands available in the CLI, along with their associated flags.

#### `run` Command

This command initiates the Galadriel server.

```bash
./galadriel-server run [flags]
```

| Flag           | Description                               | Default                   |
|----------------|-------------------------------------------|---------------------------|
| `-c, --config` | Path to the Galadriel Server config file. | `conf/server/server.conf` |

#### `token generate` Command

This 'generate' command enables the generation of a join token bound to the provided trust domain. The join token acts
as a secure authentication mechanism to establish the requisite trust relationship between the Harvester and the
Galadriel Server.

```bash
./galadriel-server token generate [flags]
```

| Flag                | Description                                             | Default |
|---------------------|---------------------------------------------------------|---------|
| `-t, --trustDomain` | The trust domain to which the join token will be bound. |         |
| `--ttl`             | Token TTL in seconds.                                   | `600`   |

#### `trustdomain` Command

The 'trustdomain' command facilitates the management of SPIFFE trust domains in the Galadriel Server. This
command allows for the registration, listing, updating, and deletion of trust domains.

```bash
./galadriel-server trustdomain [command]
```

Subcommands:

- `create`: Register a new trust domain in Galadriel Server.

##### `trustdomain create` Subcommand

This 'create' command registers a new trust domain in the Galadriel Server.

```bash
./galadriel-server trustdomain create [flags]
```

| Flag                | Description                               | Default |
|---------------------|-------------------------------------------|---------|
| `-t, --trustDomain` | The name of the trust domain to register. |         |

#### `relationship` Command

The 'relationship' command manages federation relationships between SPIFFE trust domains. Federation relationships in
SPIFFE enable secure communication between workloads across different trust domains.

```bash
./galadriel-server relationship [command]
```

Subcommands:

- `create`: Register a new federation relationship in Galadriel Server.

##### `relationship create` Subcommand

This 'create' command registers a new federation relationship in the Galadriel Server.

```bash
./galadriel-server relationship create [flags]
```

| Flag                 | Description                                                    | Default |
|----------------------|----------------------------------------------------------------|---------|
| `-a, --trustDomainA` | The name of a trust domain to participate in the relationship. |         |
| `-b, --trustDomainB` | The name of a trust domain to participate in the relationship. |         |

### Global Flags

These flags can be used across all commands.

| Flag           | Description                              | Default                          |
|----------------|------------------------------------------|----------------------------------|
| `--socketPath` | Path to the Galadriel Server API socket. | `/tmp/galadriel-server/api.sock` |

## Sample Configuration File

The following is a sample configuration file for the Galadriel server:

```hcl
server {
  listen_address = "localhost"
  listen_port = "8085"
  socket_path = "/tmp/galadriel-server/api.sock"
  log_level = "DEBUG"
}

providers {
  Datastore "sqlite3" {
    connection_string = "./datastore.sqlite3"
  }

  X509CA "disk" {
    key_file_path = "./conf/server/dummy_root_ca.key"
    cert_file_path = "./conf/server/dummy_root_ca.crt"
  }

  KeyManager "memory" {}
}
```
