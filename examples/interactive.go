package main

import (
	"fmt"
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/topicusonderwijs/go-slackbot/pkg/slackbot"
	"net/http"
	"time"
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
	// Start GC to delete old callbacks that are created more that an hour ago.
	go slackbot.GCCallback(15 * time.Minute)

	bot.SetHTTPHandleFunctions(mux)
	bot.RegisterCallbackEvent(slackevents.AppMention, AppMentionEvent)

	bot.RegisterInteractionCallback(slack.InteractionTypeBlockActions, "page_back", ActionShowPrev)
	bot.RegisterInteractionCallback(slack.InteractionTypeBlockActions, "page_forward", ActionShowNext)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error while serving: %s", err)
	}

}

func genBlocks(callback *slackbot.Callback) []slack.Block {

	data := []string{"Line1", "Line2", "Line3", "Line4", "Line5", "Line6", "Line7"}
	maxPerPage := callback.GetInt("maxPerPage")
	count := len(data)

	divSection := slack.NewDividerBlock()
	blocks := make([]slack.Block, 0)

	headerText := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Found %d item(s)", count), false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	blocks = append(blocks, headerSection)

	start := callback.GetInt("start")
	end := callback.GetInt("end")

	if end > count {
		end = count
	}

	for i := start; i < end; i++ {
		blocks = append(blocks, divSection)
		blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", data[i], false, false), nil, nil))
	}

	if count >= maxPerPage {

		buttonBlocks := make([]slack.BlockElement, 0)

		// Add a next button:
		PageForward := slack.NewButtonBlockElement(
			callback.AddUUID().String(),
			"page_forward",
			slack.NewTextBlockObject(
				"plain_text",
				fmt.Sprintf("Next %d", maxPerPage),
				true,
				false,
			),
		)

		// Add a previous button:
		PageBackward := slack.NewButtonBlockElement(
			callback.AddUUID().String(),
			"page_back",
			slack.NewTextBlockObject(
				"plain_text",
				fmt.Sprintf("Previous %d", maxPerPage),
				true,
				false,
			),
		)

		log.Debugf("Numbers: Start: %d, End: %d, Count: %d ", start, end, count)
		if count > maxPerPage {
			if start > 0 {
				buttonBlocks = append(buttonBlocks, *PageBackward)
			}
			if start < (count - maxPerPage) {
				buttonBlocks = append(buttonBlocks, *PageForward)
			}
		}

		if len(buttonBlocks) > 0 {
			actionsBlock := slack.NewActionBlock(
				"",
				buttonBlocks...,
			)

			blocks = append(blocks, actionsBlock)
			headerText.Text = fmt.Sprintf("%s (Showing %d to %d)", headerText.Text, start+1, end)
		}
	}
	return blocks
}

func ActionShowNext(payload slack.InteractionCallback, ctx *slackbot.Context) (retMsg slack.Message) {
	return ActionShowMore(1, payload, ctx)
}
func ActionShowPrev(payload slack.InteractionCallback, ctx *slackbot.Context) (retMsg slack.Message) {
	return ActionShowMore(-1, payload, ctx)
}

func ActionShowMore(offset int, payload slack.InteractionCallback, ctx *slackbot.Context) (retMsg slack.Message) {

	callback, err := slackbot.FindCallback(payload.ActionCallback.BlockActions[0].ActionID)
	if err != nil {
		log.Errorf("Callback id not found: %s", payload.ActionCallback.BlockActions[0].ActionID)
		return
	}

	start := callback.GetInt("start")
	end := callback.GetInt("end")
	perpage := callback.GetInt("maxPerPage")

	if offset > 0 {
		start = start + perpage
		end = end + perpage
	} else {
		start = start - perpage
		end = end - perpage
		if start < 0 {
			start = 0
			end = perpage
		}
	}

	callback.Set("start", start)
	callback.Set("end", end)

	blocks := genBlocks(callback)

	_, err = ctx.Api.PostEphemeral(
		callback.GetString("channel"),
		callback.GetString("user"),
		slack.MsgOptionReplaceOriginal(payload.ResponseURL),
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		log.Error(err.Error())
	}

	return
}

func AppMentionEvent(event slackevents.EventsAPIEvent, ctx *slackbot.Context) {
	ev := event.InnerEvent.Data.(*slackevents.AppMentionEvent)

	callback := slackbot.NewCallback()
	callback.Set("clicked", 0)
	callback.Set("start", 0)
	callback.Set("end", 2)
	callback.Set("maxPerPage", 3)
	callback.Set("channel", ev.Channel)
	callback.Set("user", ev.User)

	blocks := genBlocks(callback)

	str, err := ctx.Api.PostEphemeral(
		ev.Channel,
		ev.User,
		slack.MsgOptionPostEphemeral(ev.User),
		slack.MsgOptionBlocks(blocks...),
	)

	if err != nil {
		fmt.Printf("failed posting message: %v", err)
	}
	fmt.Println("return string: ", str)
}
