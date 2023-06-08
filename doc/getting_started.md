# Getting Started with Galadriel

This guide will walk you through the process of setting up Galadriel, a tool that streamlines the configuration of
Federation relationships among SPIRE Servers and manages the secure exchange of Trust Bundles.

## Prerequisites

Before you begin, you need to have two SPIRE Servers running in two different trust domains: `trust-domain-a`
and `trust-domain-b`. Each SPIRE Server should bind the local API to a unique UNIX Domain Socket (UDS). For more
information on how to set up a SPIRE Server, refer to
the [SPIRE Server Getting Started Guide](https://spiffe.io/docs/latest/try/getting-started-linux-macos-x/).

For example:

- SPIRE Server for `trust-domain-a` should bind to `/tmp/spire-server-a/private/api.sock`
- SPIRE Server for `trust-domain-b` should bind to `/tmp/spire-server-b/private/api.sock`

## Installing Galadriel

You can either download the pre-compiled binaries for Linux or build Galadriel from source for MacOS.

### Downloading Galadriel for Linux

You can download the Galadriel binaries for `amd64` and `arm64` architectures using the following commands:

For `amd64` architecture:

```bash
curl -s -N -L https://github.com/HewlettPackard/galadriel/releases/download/v0.2.1/galadriel-0.2.1-linux-amd64-glibc.tar.gz | tar xz
```

For `arm64` architecture:

```bash
curl -s -N -L https://github.com/HewlettPackard/galadriel/releases/download/v0.2.1/galadriel-0.2.1-linux-arm64-glibc.tar.gz | tar xz
```

These commands will download and extract the Galadriel binaries `galadriel-server` and `galadriel-harvester` into a
directory named `galadriel-0.2.1`. The binaries can be found in the `bin` directory.

### Building Galadriel from Source for MacOS

To build Galadriel from source, clone the Galadriel repository and build the binaries using `make`:

```bash
git clone https://github.com/HewlettPackard/galadriel.git
cd galadriel
make build
```

This will create the Galadriel binaries `galadriel-server` and `galadriel-harvester` in the `bin` directory.

## Starting the Galadriel Server

The Galadriel Server is responsible for managing relationships between trust domains and for storing and distributing
the trust bundles to the Harvesters.

To start the Galadriel Server, navigate to the root directory of the Galadriel release artifact and run the following
command:

```bash
./bin/galadriel-server run --config conf/server/server.conf
```

This command will start the Galadriel Server using the configuration file located at `conf/server/server.conf`.

## Registering Trust Domains in Galadriel Server

You need to register the trust domains `trust-domain-a` and `trust-domain-b` in the Galadriel Server. You can do this
using the following commands:

```bash
./galadriel-server trustdomain create --trustDomain trust-domain-a
```

```bash
./galadriel-server trustdomain create --trustDomain trust-domain-b
```

These commands will register the trust domains and output a confirmation message.

## Creating a Relationship Between Trust Domains

To create a relationship between the trust domains `trust-domain-a` and `trust-domain-b`, use the following command:

```bash
./galadriel-server relationship create --trustDomainA trust-domain-a --trustDomainB trust-domain-b
```

This command creates a relationship between the trust domains

`trust-domain-a` and `trust-domain-b` that is initially in the `PENDING` state. This relationship needs to be approved
by the Harvesters administrators.

## Generating a Join Token to Onboard a Harvester

A join token is required to onboard a Harvester. To generate a join token bound to trust domain `trust-domain-a`, run
the following command:

```bash
./galadriel-server token generate --trustDomain trust-domain-a
```

This command will output a token string.

## Starting the First Harvester

The Harvester is responsible for fetching and uploading trust bundles from/to the Galadriel Server and setting them to
the SPIRE Server.

### Harvester Config File

First, copy the sample Harvester config file `conf/harvester/harvester.conf` to `conf/harvester/harvester-a.conf` and
edit the following properties:

```hcl
harvester {
  trust_domain = "trust-domain-a"
  harvester_socket_path = "/tmp/galadriel-harvester-a/api.sock"
  spire_socket_path = "/tmp/spire-server-a/private/api.sock"
}
```

To start the Harvester, run the following command, replacing `<token_string>` with the token generated in the previous
step:

```bash
./galadriel-harvester run --joinToken <token_string> --config conf/harvester/harvester-a.conf
```

## Starting the Second Harvester

To start a second Harvester, you need to generate a join token bound to trust domain `trust-domain-b`:

```bash
./galadriel-server token generate --trustDomain trust-domain-b
```

Next, copy the sample Harvester config file `conf/harvester/harvester.conf` to `conf/harvester/harvester-b.conf` and
edit the following properties:

```hcl
harvester {
  trust_domain = "trust-domain-b"
  harvester_socket_path = "/tmp/galadriel-harvester-b/api.sock"
  spire_socket_path = "/tmp/spire-server-b/private/api.sock"
}
```

Start the second Harvester by running the following command, replacing `<token_string>` with the token generated in the
previous step:

```bash
./galadriel-harvester run --joinToken <token_string> --config conf/harvester/harvester-b.conf
```

## Approving the Relationship Between Trust Domains

To list the relationships, run the following command:

```bash
./galadriel-harvester relationship list --socketPath /tmp/galadriel-harvester-a/api.sock
```

This command will output the details of the relationship, including its ID and the consent status of each trust domain.

To approve the relationship, run the following command, replacing `<relationship-id>` with the relationship ID:

```bash
./galadriel-harvester relationship approve --socketPath /tmp/galadriel-harvester-a/api.sock --relationshipID <relationship-id>
```

This command will approve the relationship from the perspective of `trust-domain-a`. Now, the trust bundles
from `trust-domain-b` will be fetched from the Galadriel Server and set into the SPIRE Server.

Repeat the approval process for the `trust-domain-b` Harvester:

```bash
./galadriel-harvester relationship approve --socketPath /tmp/galadriel-harvester-b/api.sock --relationshipID <relationship-id>
```

Now the relationship is approved by both Harvesters, which means that the trust bundles from one trust domain will be
fetched from the Galadriel Server and set into the SPIRE Server in the other trust domain.

## Verifying the Bundles Were Set in the SPIRE Servers

In the SPIRE Servers, you should see a log line similar to the following:

```bash
INFO[0082] Bundle set successfully                       authorized_as=local authorized_via=transport method=BatchSetFederatedBundle trust_domain_id=trust_domain_a
```

This indicates that the trust bundles have been successfully set.

To verify the Federated bundles in the SPIRE Servers, use the following commands:

```bash
./spire-server bundle list -socketPath /tmp/spire-server-a/private/api.sock
```

```bash
./spire-server bundle list -socketPath /tmp/spire-server-b/private/api.sock
```

These commands will display the Federated bundles for each trust domain.

Congratulations! You have now set up Galadriel and established a Federation relationship between two trust domains.