package slackbot

import (
	"fmt"
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"net/http"
	"strings"
)

type SlackBot struct {
	config struct {
		signSecret string
		botToken   string
		appToken   string
	}

	registeredCommands  map[string]CommandFunc
	registeredCallbacks map[slack.InteractionType]map[string]InteractionCallbackFunc
	registeredEvents    map[string]CallbackEventFunc

	Api          *slack.Client
	socketClient *socketmode.Client
}

func NewSlackBot(slackSignSecret, slackBotToken, slackAppToken string) *SlackBot {

	slackBot := SlackBot{}

	slackBot.config.signSecret = slackSignSecret
	slackBot.config.botToken = slackBotToken
	slackBot.config.appToken = slackAppToken

	slackBot.Setup()

	return &slackBot

}

func (s *SlackBot) SetHTTPHandleFunctions(http *http.ServeMux) {

	http.HandleFunc("/events", s.DefaultHandler)
	http.HandleFunc("/slack/events", s.EventsHandler)
	http.HandleFunc("/slack/load-options", s.DefaultHandler)

	http.HandleFunc("/slack/actions", s.ActionsHandler)
	http.HandleFunc("/slack/commands", s.CommandsHandler)

}

type CommandFunc func(command slack.SlashCommand) slack.Message
type InteractionCallbackFunc func(callback slack.InteractionCallback) slack.Message
type CallbackEventFunc func(event slackevents.EventsAPIEvent)

func (s *SlackBot) RegisterCommand(command string, handler CommandFunc) error {

	if strings.HasPrefix(command, "/") == false {
		return fmt.Errorf("command should start with a /")
	}

	if _, ok := s.registeredCommands[command]; ok != false {
		return fmt.Errorf("command '%s' allready registered", command)
	}

	log.Debugf("Registering command: %s", command)
	s.registeredCommands[command] = handler

	return nil

}

func (s *SlackBot) RegisterInteractionCallback(interactionType slack.InteractionType, callbackId string, handler InteractionCallbackFunc) error {

	if _, ok := s.registeredCallbacks[interactionType][callbackId]; ok != false {
		return fmt.Errorf("%s Callback '%s' allready registered", interactionType, callbackId)
	}

	if _, ok := s.registeredCallbacks[interactionType]; ok == false {
		s.registeredCallbacks[interactionType] = make(map[string]InteractionCallbackFunc)
	}

	log.Debugf("Registering callbackId: %s", callbackId)
	s.registeredCallbacks[interactionType][callbackId] = handler

	return nil

}

func (s *SlackBot) RegisterCallbackEvent(event string, handler CallbackEventFunc) error {

	if _, ok := s.registeredEvents[event]; ok != false {
		return fmt.Errorf("event '%s' allready registered", event)
	}

	log.Debugf("Registering event: %s", event)
	s.registeredEvents[event] = handler

	return nil

}

/** SlackBot Methods **/

func (s *SlackBot) Setup() {

	s.registeredCommands = make(map[string]CommandFunc)
	s.registeredCallbacks = make(map[slack.InteractionType]map[string]InteractionCallbackFunc)
	s.registeredEvents = make(map[string]CallbackEventFunc)

	if s.Api == nil && s.config.botToken != "" {
		s.Api = slack.New(
			s.config.botToken,
			slack.OptionDebug(false),
			slack.OptionAppLevelToken(s.config.appToken),
		)
		s.socketClient = socketmode.New(
			s.Api,
			socketmode.OptionDebug(false),
		)
	}

	s.StartSocketListener()

}

func (s *SlackBot) FireCommand(command slack.SlashCommand) slack.Message {

	var payload slack.Message

	if commandFunc, ok := s.registeredCommands[command.Command]; ok != false {
		log.Debugln(command.Command, " found")
		payload = commandFunc(command)
	} else {
		payload.Msg = slack.Msg{Text: fmt.Sprintf("Unknown command: %s %s", command.Command, command.Text)}
	}
	return payload

}

func (s *SlackBot) FireInteractiveCallback(interactionCallback slack.InteractionCallback) slack.Message {

	var payload slack.Message

	callbackId := interactionCallback.CallbackID

	if interactionCallback.Type == slack.InteractionTypeViewSubmission {
		callbackId = interactionCallback.View.CallbackID
	}

	if interactionCallback.Type == slack.InteractionTypeBlockActions {
		callbackId = interactionCallback.Value

		if len(interactionCallback.ActionCallback.BlockActions) > 0 {
			log.Debugln("BlockActions")
			callbackId = interactionCallback.ActionCallback.BlockActions[0].Value
			if callbackId == "" {
				callbackId = interactionCallback.ActionCallback.BlockActions[0].SelectedOption.Value
			}
		}

		if len(interactionCallback.ActionCallback.AttachmentActions) > 0 {
			log.Debugln("AttachmentActions")
			callbackId = interactionCallback.ActionCallback.AttachmentActions[0].Value
		}

	}

	if callbacks, ok := s.registeredCallbacks[interactionCallback.Type]; ok != false {

		if callbackFunc, ok := callbacks[callbackId]; ok != false {
			log.Debugf("Callback %s found", callbackId)
			callbackFunc(interactionCallback)
		} else {
			log.Debugf("Callback %s not found", callbackId)
		}

	} else {
		log.Debugf("Unknown callback: %s", callbackId)
		payload.Msg = slack.Msg{Text: fmt.Sprintf("Unknown callback: %s", callbackId)}
	}

	/*
		jout, err := json.MarshalIndent(interactionCallback, "", "    ")
		if err != nil {
			log.Error(err)
		}
		log.Debugf("%s", jout)
	*/
	return payload

}

func (s *SlackBot) FireCallbackEvent(eventsAPIEvent slackevents.EventsAPIEvent) {

	innerEvent := eventsAPIEvent.InnerEvent

	eventType := innerEvent.Type

	if eventFunc, ok := s.registeredEvents[eventType]; ok != false {
		eventFunc(eventsAPIEvent)
	} else {
		log.Debugf("Event %s not registered", eventType)
	}

}
