package slackbot

import (
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func (s *SlackBot) SocketListener() {
	for evt := range s.socketClient.Events {
		log.Debugln("Got event")
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			log.Traceln("Connecting to Slack with Socket Mode...")
		case socketmode.EventTypeConnectionError:
			log.Traceln("Connection failed. Retrying later...")
		case socketmode.EventTypeConnected:
			log.Traceln("Connected to Slack with Socket Mode.")
		case socketmode.EventTypeEventsAPI:
			eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
			if !ok {
				log.Warnf("Ignored %+v\n", evt)
				continue
			}
			log.Debugf("Event received: %+v\n", eventsAPIEvent)

			s.socketClient.Ack(*evt.Request)

			switch eventsAPIEvent.Type {
			case slackevents.CallbackEvent:
				s.FireCallbackEvent(eventsAPIEvent)
			case slackevents.URLVerification:
				 log.Warnln("Url Verification event received")
			case slackevents.AppRateLimited:
				// AppRateLimited indicates your app's event subscriptions are being rate limited
				log.Warnln("AppRateLimited event received")

			default:
				s.socketClient.Debugf("unsupported Events API event received")
			}
		case socketmode.EventTypeInteractive:
			callback, ok := evt.Data.(slack.InteractionCallback)
			if !ok {
				log.Warnf("Ignored %+v\n", evt)
				continue
			}

			var payload interface{}
			payload = s.FireInteractiveCallback(callback)
			s.socketClient.Ack(*evt.Request, payload)
		case socketmode.EventTypeSlashCommand:
			cmd, ok := evt.Data.(slack.SlashCommand)
			if !ok {
				log.Debugf("Ignored %+v\n", evt)
				continue
			}

			var payload slack.Message
			payload = s.FireCommand(cmd)
			s.socketClient.Ack(*evt.Request, payload)

		case socketmode.EventTypeHello:
			//s.socketClient.Ack(*evt.Request)
		default:
			log.Errorf("Unexpected event type received: %s\n", evt.Type)
		}
	}

}


func (s *SlackBot) StartSocketListener() {
	go s.SocketListener()

	if s.socketClient != nil {
		go s.socketClient.Run()
	} else {
		log.Warn("SocketClient is nill")
	}

}
