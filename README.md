# Galadriel
[![CodeQL](https://github.com/HewlettPackard/galadriel/actions/workflows/codeql.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/codeql.yml)
[![PR Build](https://github.com/HewlettPackard/galadriel/actions/workflows/pr_build.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/pr_build.yml)
[![Scorecards supply-chain security](https://github.com/HewlettPackard/galadriel/actions/workflows/scorecards.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/scorecards.yml)
[![trivy](https://github.com/HewlettPackard/galadriel/actions/workflows/trivy.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/trivy.yml)

---

Project Galadriel, or just Galadriel, is an open source project that enables scalable and easy configuration of Federation relationships among SPIRE Servers. It works as a central hub for managing and auditing Federation relationships.

### What Galadriel IS?
- **Alternative approach to SPIRE Federation**: it's built on top of SPIRE APIs to facilitate foreign Trust Bundles management.
- **Multi-tenant**: multiple organizations can leverage the same Galadriel deployment, while ensuring data and operations isolation.
- **Federation at scale**: configuring multiple SPIRE Server federation should be easy and secure by default, that is Galadriel's main focus.
- **Central hub**: it's a central place where federation relationships can be defined and audited.

### What Galadriel IS NOT?
- **A replacement of SPIRE/SPIFFE Federation**: it doesn't replace SPIRE Federation, it leverages what's already built in there.
- **A SPIRE plugin**: it's deployed as a separate component, not as a SPIRE plugin.

---

## Get started

- Learn how to run the Proof of Concept (v0.1.0) [here](https://github.com/HewlettPackard/galadriel/blob/v0.1.0/doc/INSTRUCTIONS.md)
- [Configuration and CLI Usage instructions](./doc/USAGE.md)

## Contribute

Project Galadriel is an open source project under the [Apache 2 license](./LICENSE), and as such, any kind of contribution is welcome, being documentation, new features, bugfixing, issues, etc. Check out our [Contributing guidelines](./CONTRIBUTING.md) to learn how we manage contributions, and the [Governance policy](./GOVERNANCE.md) to learn about the different roles in the project.

## Roadmap

Project Galadriel has currently reached the Proof of Concept milestone ([v0.1.0](https://github.com/HewlettPackard/galadriel/blob/v0.1.0/doc/INSTRUCTIONS.md)). Refer to the [Roadmap](./ROADMAP.md) to learn what's next.

## Want to know more?

### Design document
Please feel free to check out our [Design Document](https://docs.google.com/document/d/1nkiJV4PAV8Wx1oNvx4CT3IDtDRvUFSL8/edit?usp=sharing&ouid=106690422347586185642&rtpof=true&sd=true), where you can find more information about the architecture and future plans for Galadriel. Comments and suggestions are welcome and highly appreciated.

### Community Presentations & Blog Posts
- SPIRE Bridge: an Alternative Approach to SPIFFE Federation - [Juliano Fantozzi](https://github.com/jufantozzi), [Maximiliano Churichi](https://github.com/mchurichi) / SPIFFE Community Day Fall 2022 (October 2022) / [video](https://www.youtube.com/watch?v=pHdOm4MdPHE), [slides](https://docs.google.com/presentation/d/1Cox9MNeZA1bD2aktg2HTMjcgGn_6Rbb0/edit?usp=sharing&ouid=106690422347586185642&rtpof=true&sd=true), [demo](https://github.com/HewlettPackard/galadriel/tree/v0.1.0/demos)
- Galadriel - A SPIRE Federation Alternative - [William Barrera Fuentes](https://github.com/wibarre) / HPE Developer Community (October 2022) / [blog post](https://developer.hpe.com/blog/galadriel-a-spire-federation-alternative/)

## Found a security issue?

Please refer to the [Security policy](./SECURITY.md) to learn more about security updates and reporting potential vulnerabilities.
