# Galadriel

[![CodeQL](https://github.com/HewlettPackard/galadriel/actions/workflows/codeql.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/codeql.yml)
[![PR Build](https://github.com/HewlettPackard/galadriel/actions/workflows/pr_build.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/pr_build.yml)
[![Scorecards supply-chain security](https://github.com/HewlettPackard/galadriel/actions/workflows/scorecards.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/scorecards.yml)
[![trivy](https://github.com/HewlettPackard/galadriel/actions/workflows/trivy.yml/badge.svg)](https://github.com/HewlettPackard/galadriel/actions/workflows/trivy.yml)

---

Project Galadriel is an open-source project that streamlines the configuration of Federation relationships among SPIRE
Servers and manages the secure exchange of Trust Bundles based on the registered and approved relationships. It
functions as a central hub for the management and auditing of these Federation relationships.

### What is Galadriel?

- **Alternative approach to SPIRE Federation**: Galadriel is built on top of SPIRE APIs to streamline the management of
  foreign Trust Bundles.
- **Federation at scale**: Galadriel simplifies the configuration of multiple SPIRE Server federations while
  prioritizing security.
- **Central hub**: Galadriel provides a centralized platform where federation relationships can be defined and audited.

### What Galadriel is not?

- **A replacement for SPIRE/SPIFFE Federation**: Galadriel does not replace SPIRE Federation, instead, it leverages
  existing SPIRE capabilities.
- **A SPIRE plugin**: Galadriel is deployed as a standalone component, not as a SPIRE plugin.

---

## Getting Started

- **TBD**

## Contributing

Project Galadriel is an open-source project under the [Apache 2 license](./LICENSE). We welcome any form of
contribution, whether it's documentation, new features, bug fixes, or issues. Check out
our [Contributing guidelines](./CONTRIBUTING.md) to learn about our contribution management, and
the [Governance policy](./GOVERNANCE.md) to understand the various roles within the project.

## Roadmap

Project Galadriel has currently reached the Proof of Concept
milestone ([v0.1.0](https://github.com/HewlettPackard/galadriel/blob/v0.1.0/doc/INSTRUCTIONS.md)). Refer to
the [Roadmap](./ROADMAP.md) to learn about our future plans.

## Want to Know More?

### Design Document

Feel free to explore
our [Design Document](https://docs.google.com/document/d/1nkiJV4PAV8Wx1oNvx4CT3IDtDRvUFSL8/edit?usp=sharing&ouid=106690422347586185642&rtpof=true&sd=true),
which provides more information about Galadriel's architecture and future plans. Your comments and suggestions are
welcome and highly appreciated.

### Community Presentations & Blog Posts

- SPIRE Bridge: an Alternative Approach to SPIFFE
  Federation - [Juliano Fantozzi](https://github.com/jufantozzi), [Maximiliano Churichi](https://github.com/mchurichi) /
  SPIFFE Community Day Fall 2022 (October
    2022) / [video](https://www.youtube.com/watch?v=pHdOm4MdPHE), [slides](https://docs.google.com/presentation/d/1Cox9MNeZA1bD2aktg2HTMjcgGn_6Rbb0/edit?usp=sharing&ouid=106690422347586185642&rtpof=true&sd=true), [demo](https://github.com/HewlettPackard/galadriel/tree/v0.1.0/demos)
- Galadriel - A SPIRE Federation Alternative - [William Barrera Fuentes](https://github.com/wibarre) / HPE Developer
  Community (October 2022) / [blog post](https://developer.hpe.com/blog/galadriel-a-spire-federation-alternative/)

## Encountered a Security Issue?

Please refer to our [Security policy](./SECURITY.md) for more information about security updates and how to report
potential vulnerabilities.
