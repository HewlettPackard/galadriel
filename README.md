# Galadriel

[![CodeQL](https://github.com/HewlettPackard/galadriel/actions/workflows/codeql.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/codeql.yml)
[![PR Build](https://github.com/HewlettPackard/galadriel/actions/workflows/pr_build.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/pr_build.yml)
[![Scorecards supply-chain security](https://github.com/HewlettPackard/galadriel/actions/workflows/scorecards.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/scorecards.yml)
[![trivy](https://github.com/HewlettPackard/galadriel/actions/workflows/trivy.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/trivy.yml)

---

Galadriel is an open-source project that enables scalable and easy configuration of Federation relationships among SPIRE Servers. It serves as a central hub for managing and auditing Federation relationships.

### What Galadriel IS?

- **Alternative approach to SPIRE Federation**: Galadriel is built on top of SPIRE APIs to facilitate foreign Trust Bundles management.
- **Federation at scale**: The main focus of Galadriel is to make the configuration of multiple SPIRE Server federations easy and secure by default.
- **Central hub**:  Galadriel provides a centralized location for defining and auditing federation relationships. It also securely stores and manages trust bundles.

### What Galadriel IS NOT?

- **A replacement of SPIRE/SPIFFE Federation**: Galadriel does not replace SPIRE Federation but leverages the existing functionality.
- **A SPIRE plugin**: Galadriel is deployed as a separate component and not as a SPIRE plugin.

---

## Get started

- TBD

## Contribute

Galadriel is an open-source project licensed under the [Apache 2 license](./LICENSE). We welcome any kind of
contribution, including documentation, new features, bug fixing, and reporting issues. Please refer to
our [Contributing guidelines](./CONTRIBUTING.md) to learn how to contribute, and
the [Governance policy](./GOVERNANCE.md) to understand the different roles in the project.

## Roadmap

Project Galadriel has currently reached the Proof of Concept milestone ([v0.1.0](https://github.com/HewlettPackard/galadriel/blob/v0.1.0/doc/INSTRUCTIONS.md)). Refer to
the [Roadmap](./ROADMAP.md) to learn what's next.

## Want to know more?

### Design document

Please check out our [Design Document](https://docs.google.com/document/d/1nkiJV4PAV8Wx1oNvx4CT3IDtDRvUFSL8/edit?usp=sharing&ouid=106690422347586185642&rtpof=true&sd=true) for more information about the architecture and future plans for Galadriel. We highly appreciate any comments and suggestions.

### Community Presentations & Blog Posts
- SPIRE Bridge: an Alternative Approach to SPIFFE Federation - [Juliano Fantozzi](https://github.com/jufantozzi), [Maximiliano Churichi](https://github.com/mchurichi) / SPIFFE Community Day Fall 2022 (October 2022) / [video](https://www.youtube.com/watch?v=pHdOm4MdPHE), [slides](https://docs.google.com/presentation/d/1Cox9MNeZA1bD2aktg2HTMjcgGn_6Rbb0/edit?usp=sharing&ouid=106690422347586185642&rtpof=true&sd=true), [demo](https://github.com/HewlettPackard/galadriel/tree/v0.1.0/demos)
- Galadriel - A SPIRE Federation Alternative - [William Barrera Fuentes](https://github.com/wibarre) / HPE Developer Community (October 2022) / [blog post](https://developer.hpe.com/blog/galadriel-a-spire-federation-alternative/)

## Found a security issue?

Please refer to the [Security policy](./SECURITY.md) to learn more about security updates and reporting potential vulnerabilities.
