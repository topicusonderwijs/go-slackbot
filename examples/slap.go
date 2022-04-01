package main

import (
	"fmt"
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/topicusonderwijs/go-slackbot/pkg/slackbot"
	"net/http"
)

func main() {

	log.EnableLevel("debug")

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}
	bot := slackbot.NewSlackBot(
		"SigningSecret",
		"BotToken",
		"AppLevelToken",
	)
	bot.SetHTTPHandleFunctions(mux)
	bot.RegisterCallbackEvent(slackevents.AppMention, AppMentionEvent)
	bot.RegisterCommand("/hello", CommandHello)
	bot.RegisterCommand("/slap", CommandSlap)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error while serving: %s", err)
	}

}

func AppMentionEvent(event slackevents.EventsAPIEvent, ctx *slackbot.Context) {
	ev := event.InnerEvent.Data.(*slackevents.AppMentionEvent)

	_, _, err := ctx.Api.PostMessage(ev.Channel, slack.MsgOptionText("Yes, hello.", false))
	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
}

func CommandHello(command slack.SlashCommand, ctx *slackbot.Context) (payload slack.Message) {
	// This will create an "only visible to you" message
	payload.Msg = slack.Msg{Text: "Oh hi there"}
	return
}
func CommandSlap(command slack.SlashCommand, ctx *slackbot.Context) (payload slack.Message) {
	ctx.Ack(*ctx.Event.Request)
	// Warning command.Text is returned unfiltered and unescaped and could result in unsafe/unexpected behavior.
	ctx.Api.SendMessage(command.ChannelID, slack.MsgOptionText(fmt.Sprintf("*%s* wants to slap *%s* a bit around with a big large trout", command.UserName, command.Text), false))
	return
}
