# Changelog

## [0.2.0] - 2023-06-06

### Added

- TLS communication between Galadriel Server and Harvester for enhanced security (#146).
- Enhanced Harvester secure introduction flow by utilizing join tokens and issuing JWTs by the Server for Harvester authentication (#151).
- Bundle signing and verification using generic interface and providing a `disk` implementation (#147).
- APIs for Server and Harvester defined through Open API spec for improved documentation and client integration (#70, #170).
- Harvester admin API specification (#170).
- Galadriel Server Admin API implementation (#150).
- Harvester Admin API implementation (#154).
- Datastore layer supporting SQLite and Postgres (#73, #157).
- Comprehensive overhaul of Harvester and Server, incorporating various enhancements, including improved synchronization processes for SPIRE bundles and Federated bundles (#171).
- Improvements in CLI implementations (#173).
- KeyManager `disk` implementation (#167).
- X509CA `disk` implementation for managing X.509 certificate authorities (#145).
- `diskutil` package for atomic file writing operations (#187).
- Releasing how-to document (#186).
- Improvements in build scripts and automated release process (#185, #190).
- New CLI implementations and improvements (#173, #164).


## [0.1.0] - 2022-10-31

### Added

First POC.

### Changed
### Security
