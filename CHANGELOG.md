# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Example `-socket` flag to run the slap/interactive examples in either HTTP or socket mode.
### Changed
- Upgraded to Go 1.26 and `slack-go/slack` v0.26.0.
- **Breaking Change**: `RegisterCallbackEvent` now takes a `slackevents.EventsAPIType` instead of a `string`.
- **Breaking Change**: `NewSlackBot` no longer starts socket mode automatically. Call the blocking `RunSocket()` yourself after registering handlers.
- **Breaking Change**: `RunSocket` now returns an `error` instead of calling `log.Fatalf`, so the caller decides how to handle a socket failure.
- `CallbackStorage` is now a `sync.Map` and each `Callback` guards its storage with a mutex, making the callback store concurrency-safe.
### Deprecated
### Removed
- **Breaking Change**: `StartSocketListener` was removed; its role is now covered by `RunSocket`.
### Fixed
- `GCCallback` now expires callbacks correctly and sweeps repeatedly instead of running only once.
- HTTP `ActionsHandler` no longer always returns HTTP 500; `EventsHandler` now parses and dispatches callback events and returns 200.
- Interactive callback handlers' return value is now used for the ack/HTTP response.
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