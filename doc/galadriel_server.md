# Galadriel Server Configuration Reference

This document provides a reference for the Galadriel Server configuration file.

## Configuration File

The Galadriel Server configuration file contains several sections that allow you to customize the behavior of the
Galadriel Server.

### `server`

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

### `providers`

| Provider     | Description                                                                  |
|--------------|------------------------------------------------------------------------------|
| `Datastore`  | Configures the datastore provider.                                           |
| `X509CA`     | Configures the X509CA provider for signing TLS X.509 certificates.           |
| `KeyManager` | Configures the KeyManager for providing private keys for signing JWT tokens. |

#### Datastore

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

#### X509CA

| Option | Description                                                                                                                                                                                                                |
|--------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `disk` | Uses a ROOT CA loaded from disk to issue X509 certificates. The `key_file_path` is the path to the root CA private key file in PEM format. The `cert_file_path` is the path to the root CA certificate file in PEM format. |

#### Example:

```hcl
providers {
  X509CA "disk" {
    key_file_path = "./conf/server/dummy_root_ca.key"
    cert_file_path = "./conf/server/dummy_root_ca.crt"
  }
}
```

#### KeyManager

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
```

## Galadriel Server CLI Reference

The Galadriel server provides a command-line interface (CLI) for running the server, managing federation relationships,
creating join tokens, and managing SPIFFE trust domains.

To access the CLI, use the `galadriel-server` command:

```bash
./galadriel-server
```

### Available Commands

#### `run`

Run this command to start the Galadriel server.

```bash
./galadriel-server run [flags]
```

| Flag           | Description                               | Default                   |
|----------------|-------------------------------------------|---------------------------|
| `-c, --config` | Path to the Galadriel Server config file. | `conf/server/server.conf` |

#### `token generate`

The 'generate' command allows you to generate a join token for the provided trust domain. This join token serves as a
secure authentication mechanism to establish the necessary trust relationship between the Harvester and the Galadriel
Server.

```bash
./galadriel-server token generate [flags]
```

| Flag                | Description                                             | Default |
|---------------------|---------------------------------------------------------|---------|
| `-t, --trustDomain` | The trust domain to which the join token will be bound. |         |
| `--ttl`             | Token TTL in seconds.                                   | `600`   |

#### `trustdomain`

The 'trustdomain' command is used for managing SPIFFE trust domains in the Galadriel Server database. It allows you to
register, list, update, and delete trust domains.

```bash
./galadriel-server trustdomain [command]
```

##### Available subcommands:

- `create`: Register a new trust domain in Galadriel Server.

##### `trustdomain create`

The `create` command registers a new trust domain in the Galadriel Server.

```bash

Syntax:

```bash
./galadriel-harvester trustdomain create [flags]
```

Example Usage:

```bash
./galadriel-harvester trustdomain create --trustDomain <trustDomainName>
```

| Flag                | Description                               | Default |
|---------------------|-------------------------------------------|---------|
| `-t, --trustDomain` | The name of the trust domain to register. |         |

#### `relationship`

Manage federation relationships between SPIFFE trust domains with the 'relationship' command. Federation relationships
in SPIFFE permit secure communication between workloads across different trust domains.

```bash
./galadriel-server relationship [command]
```

##### Available subcommands:

- `create`: Register a new Federation relationship in Galadriel Server.

##### `relationship create`

The `create` command registers a new Federation relationship in Galadriel Server.

```bash

Syntax:

```bash
./galadriel-harvester relationship create [flags]
```

Example Usage:

```bash
./galadriel-harvester relationship create --trustDomainA <trustDomainName> --trustDomainB <trustDomainName>
```

| Flag                 | Description                                                    | Default |
|----------------------|----------------------------------------------------------------|---------|
| `-a, --trustDomainA` | The name of a trust domain to participate in the relationship. |         |
| `-b, --trustDomainB` | The name of a trust domain to participate in the relationship. |         |

### Global Flags

| Flag           | Description                              | Default                          |
|----------------|------------------------------------------|----------------------------------|
| `--socketPath` | Path to the Galadriel Server API socket. | `/tmp/galadriel-server/api.sock` |

## Sample Configuration File

Below is a sample configuration file for the Galadriel server.

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
