package slackbot

import (
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func (s *SlackBot) SocketListener() {
	for socketEvent := range s.socket.Events {
		log.Debugln("Got event")
		socketContext := s.newSocketContext(&socketEvent)
		var payload interface{}
		var autoAck bool
		payload = nil
		autoAck = true

		switch socketEvent.Type {
		case socketmode.EventTypeConnecting:
			log.Traceln("Connecting to Slack with socket Mode...")
		case socketmode.EventTypeConnectionError:
			log.Traceln("Connection failed. Retrying later...")
		case socketmode.EventTypeConnected:
			log.Traceln("Connected to Slack with socket Mode.")
		case socketmode.EventTypeEventsAPI:
			var eventsAPIEvent slackevents.EventsAPIEvent
			eventsAPIEvent, ok := socketEvent.Data.(slackevents.EventsAPIEvent)
			if !ok {
				continue
			}
			log.Debugf("Event received: %+v\n", eventsAPIEvent)

			switch eventsAPIEvent.Type {
			case slackevents.CallbackEvent:
				s.FireCallbackEvent(eventsAPIEvent, socketContext)
				autoAck = true
			case slackevents.URLVerification:
				log.Warnln("Url Verification event received")
			case slackevents.AppRateLimited:
				// AppRateLimited indicates your app's event subscriptions are being rate limited
				log.Warnln("AppRateLimited event received")
			default:
				s.socket.Debugf("unsupported Events API event received")
			}

		case socketmode.EventTypeInteractive:
			callback, ok := socketEvent.Data.(slack.InteractionCallback)
			if !ok {
				continue
			}

			autoAck = true
			payload = s.FireInteractiveCallback(callback, socketContext)

		case socketmode.EventTypeSlashCommand:
			cmd, ok := socketEvent.Data.(slack.SlashCommand)
			if !ok {
				continue
			}

			autoAck = true
			payload = s.FireSlashCommand(cmd, socketContext)

		case socketmode.EventTypeHello:
			//s.socket.Ack(*socketEvent.Request)
		default:
			log.Errorf("Unexpected event type received: %s\n", socketEvent.Type)
		}

		if autoAck && !socketContext.IsFinished() {
			s.socket.Ack(*socketEvent.Request, payload)
		}

	}

}

func (s *SlackBot) StartSocketListener() {
	go s.SocketListener()

	if s.socket != nil {
		go s.socket.Run()
	} else {
		log.Warn("SocketClient is nill")
	}

}
