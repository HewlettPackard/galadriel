# Roadmap

## Recently Completed

- APIs for Server and Harvester defined through Open API spec.
- TLS enabled between Galadriel Server and Harvesters, using a disk-based upstream CA that the Server uses to sign its
  certificate.
- Secure Harvester introduction using a single-use join token.
- Harvester authentication using JWT issued by the Server, which are rotated. The JWT is issued by the Server using
  either an in-memory KeyManager or a disk-based KeyManager for generating the private keys.
- Bundle signing and verification using a disk-based Signer and Verifier implementation.
- Added support for SQLite and Postgres.
- Simple implementation of the Federation Relationship approval flow.
- Federated bundle synchronization across Harvesters based on configured and approved relationships.

## Near-Term and Medium-Term

- Support for SPIRE running in high-availability (HA) mode.
- Support for Galadriel Server in high-availability (HA) mode.
- Support for other upstream CAs for TLS certificates.
- Support for other Key Management Systems (KMS) for the private keys used for JWT issuing.
- Support for relationship consent signing.
- Support for other bundle signers and verifiers, e.g., using Sigstore.
- Telemetry, health checkers, alerts, and API versioning.

## Long-Term

### Initial Proof of Concept (PoC)

- **Status**: Completed ([v0.1.0](https://github.com/HewlettPackard/galadriel/tree/v0.1.0/))
- **Goal**: Exercise concepts about trust bundle exchange based on relationships. This stage corroborates the
  feasibility of having a Harvester agent as a medium to manage federated relationships in SPIRE servers, the Server as
  a central hub for exchange, and the relationship as a control for the exchange.
