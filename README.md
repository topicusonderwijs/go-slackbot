# Go-Slackbot

A go slackbot framework to make implementing a slackbot a bit easier and less repetitive.

This framework serves as an extension to `github.com/slack-go/slack`.

# Installing

```bash
go get github.com/topicusonderwijs/go-slackbot
```

# Usage

# Examples

See also: [examples](examples)

```golang

import (
    "github.com/slack-go/slack/slackevents"
    "github.com/topicusonderwijs/go-slackbot/pkg/slackbot"
)

func main() {

    http := http.NewServeMux()
    server := &http.Server{Addr: s.config.port, Handler: s.http}
    bot := slackbot.NewSlackBot(
        "SigningSecret", 
        "BotToken", 
        "AppLevelToken",
    )
    bot.SetHTTPHandleFunctions(s.http)
    bot.RegisterCallbackEvent(slackevents.AppMention, AppMentionEvent)
    bot.RegisterCommand("/hello", CommandHello)

    err := bot.ListenAndServe()
    if err != nil {
        log.Fatal("Error while serving: %s", err)
    }
    
}

func AppMentionEvent(event slackevents.EventsAPIEvent, ctx slackbot.Context) {
    ev := event.InnerEvent.Data.(*slackevents.AppMentionEvent)
    _, _, err := ctx.Api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
    if err != nil {
        fmt.Printf("failed posting message: %v", err)
    }
}

func CommandHello(command slack.SlashCommand, ctx slackbot.Context) slack.Message {
    return slack.Message{Msg: slack.Msg{Text: "Oh hi there"}}	
}

```


# Contribution

Fork, edit, open a PR and we will see where we go from there 

---
*Disclaimer*: Work in Progress