# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A Go framework that wraps `github.com/slack-go/slack` to reduce boilerplate when building a Slackbot. It is a **library** (no `main`), consumed via `github.com/topicusonderwijs/go-slackbot/pkg/slackbot`. Go 1.25.

## Commands

```bash
go build ./...        # build
go vet ./...          # vet
go test ./...         # test (no tests exist yet)
gofmt -l .            # check formatting
```

## Architecture

The package `pkg/slackbot` supports **two transports for the same handler logic**, selected at construction:

- **Socket Mode** — enabled when `NewSlackBot(...)` is given a non-empty `appToken`. `Setup()` creates the `socketmode.Client` but does **not** start it; the consumer calls the blocking `RunSocket() error` (which spawns `SocketListener` and runs `socket.Run()`) and decides how to handle a failure.
- **HTTP / Events API** — used when no `appToken` is given. The consumer wires routes via `SetHTTPHandleFunctions(mux)` and runs their own `http.Server`. Routes: `/slack/events`, `/slack/actions`, `/slack/commands`.

Both transports converge on three central dispatchers in `slackbot.go`: `FireSlashCommand`, `FireInteractiveCallback`, `FireCallbackEvent`. These look up the matching handler in the registration maps (`registeredCommands`, `registeredCallbacks`, `registeredEvents`) and invoke it. Adding new behavior means registering a handler — the transport wiring is already shared.

### Context (`context.go`)
Every handler receives a `*Context` abstracting whether the call arrived over HTTP or Socket (`IsHTTP()` / `IsSocket()`). It always carries the `*slack.Client` (`ctx.Api`). Key flow control:
- `ctx.Finish()` marks the request handled so the framework skips the default response/ack. In Socket Mode, the listener auto-acks unless `IsFinished()` is true; in HTTP, `CommandsHandler` only renders the returned payload when not finished.
- `ctx.Ack(req, payload...)` acks a socket request and marks finished.

### Interaction callback routing (`FireInteractiveCallback`)
Callback ID resolution is type-dependent and non-obvious: for `ViewSubmission` it uses `View.CallbackID`; for `BlockActions` it derives the key from the first block action's `Value` (falling back to `SelectedOption.Value`) or attachment action value. Register block-action handlers under the value you set on the block element, not a literal callback ID.

### Callback storage (`callback.go`)
`Callback` is an in-memory key/value store keyed by UUID for carrying state across an interaction (e.g. between a command and a later block action). `NewCallback()` / `AddUUID()` register into the package-global `CallbackStorage`; `FindCallback(id)` retrieves (and strips surrounding quotes from the id). `GCCallback(sleep)` is the expiry sweep — call it in a loop to drop entries older than 1 hour. Note this state is process-local and not concurrency-guarded.

## Notes

- There is no `ListenAndServe()` method on `SlackBot` — the consumer owns the `http.Server` lifecycle (see `examples/` and the README).
- Logging uses `github.com/humsie/log`; debug/trace output is gated by `config.slackDebug` (no public setter currently).
- Handler signatures: `CommandFunc(slack.SlashCommand, *Context) slack.Message`, `InteractionCallbackFunc(slack.InteractionCallback, *Context) slack.Message`, `CallbackEventFunc(slackevents.EventsAPIEvent, *Context)`.
- This is a Work in Progress; CHANGELOG.md follows Keep a Changelog + SemVer.
