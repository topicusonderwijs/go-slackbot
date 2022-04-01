# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security


## [0.2.0] - 2022-03-24
### Added
- Added Context interface containing web and socket context for incoming calls
### Changed
- **Breaking Change**: Signature for registrable function changed to allow passing of context 
- **Breaking Change**: Registered functions are called with context 
- socketclient/auto-ack behavior now checks context to verify ack is still needed.


## [0.1.0] - 2022-02-10
### Added
- Initial Version