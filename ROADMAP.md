# Roadmap

## Milestones

### Proof of Concept (PoC)
- **Status**: Completed ([v0.1.0](https://github.com/HewlettPackard/galadriel/tree/v0.1.0/))
- **Goal**: Exercise concepts about trust bundle exchange based on relationships. It will corroborate the feasibility of having a Harvester agent as a medium to manage federated relationships in SPIRE servers, the Server as a middle hub for exchange, and the relationship as a control for the exchange. 
- **Result**:
    - Server runs, and stores bundles and defined relationships in an ephemeral storage system.
    - Server exposes local APIs for admins to register new members, generate access tokens for them, and define bidirectional 1:1 relationships.
    - Server exposes public authenticated APIs for Harvesters.
    - Harvester uses Server-generated access tokens to communicate with the Server.
    - Harvester communicates with the SPIRE Server to fetch its bundle and to set foreign bundles.
    - Harvester sends its collocated SPIRE bundle, and fetches and keeps in sync foreign bundles based on the defined relationships.

### Minimum Viable Product (MVP)
- **Status**: In Progress
- **Goal**: Have a minimal product for early evaluation, that is is API based, and implements the security and core principles identified in the [Design Document](https://docs.google.com/document/d/1nkiJV4PAV8Wx1oNvx4CT3IDtDRvUFSL8/edit?usp=sharing&ouid=106690422347586185642&rtpof=true&sd=true).
- **Result**:
    - Server and Harvester APIs are well defined and documented.
    - Harvester is securely introduced to the Server.
    - One or more production-ready database systems are available to be used as backend storage.
    - Multiple organizations can share the same Galadriel Server instance without data leak risks.
    - Trust bundles are cryptographically signed and verified end-to-end.
    - Galadriel supports SPIRE in an HA topology.
    - Server and Harvester can be configured to emit metrics to an open telemetry standard.
    - Harvester admins explicitly approve or deny memberships.
    - Components and flows are thoroughly and continuously tested and exercised.
    - There are deployment options for bare metal and Kubernetes.
