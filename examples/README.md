# Examples

Each example is a standalone `package main`. Run one with:

```bash
go run ./examples/slap
go run ./examples/interactive
```

Replace the placeholder `SigningSecret` / `BotToken` / `AppLevelToken` values in
each `main.go` with your own Slack app credentials before running.

| Example | Shows |
|---|---|
| [`slap`](slap) | Slash commands (`/hello`, `/slap`) and handling an `app_mention` event |
| [`interactive`](interactive) | Block actions with paging buttons, backed by the `Callback` state store and its GC |
