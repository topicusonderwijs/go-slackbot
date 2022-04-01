package slackbot

import (
	"fmt"
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"strings"
)

type SlackBot struct {
	config struct {
		signSecret string
		botToken   string
		appToken   string
		useSocket  bool
		slackDebug bool
	}

	registeredCommands  map[string]CommandFunc
	registeredCallbacks map[slack.InteractionType]map[string]InteractionCallbackFunc
	registeredEvents    map[string]CallbackEventFunc

	api    *slack.Client
	socket *socketmode.Client
}

func NewSlackBot(slackSignSecret, slackBotToken, slackAppToken string) *SlackBot {

	slackBot := SlackBot{}

	slackBot.config.signSecret = slackSignSecret
	slackBot.config.botToken = slackBotToken
	if slackAppToken != "" {
		slackBot.config.appToken = slackAppToken
		slackBot.config.useSocket = true
	} else {
		slackBot.config.useSocket = false
	}

	slackBot.Setup()

	return &slackBot

}

type CommandFunc func(command slack.SlashCommand, ctx *Context) slack.Message
type InteractionCallbackFunc func(callback slack.InteractionCallback, ctx *Context) slack.Message
type CallbackEventFunc func(event slackevents.EventsAPIEvent, ctx *Context)

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

	apiOptions := []slack.Option{}
	apiOptions = append(apiOptions, slack.OptionDebug(s.config.slackDebug))
	if s.config.useSocket {
		apiOptions = append(apiOptions, slack.OptionAppLevelToken(s.config.appToken))
	}

	if s.api == nil && s.config.botToken != "" {
		s.api = slack.New(
			s.config.botToken,
			apiOptions...,
		)
	}
	if s.api != nil && s.config.useSocket {
		s.socket = socketmode.New(
			s.api,
			socketmode.OptionDebug(s.config.slackDebug),
		)
		s.StartSocketListener()
	}
}

func (s *SlackBot) FireSlashCommand(command slack.SlashCommand, ctx *Context) slack.Message {

	var payload slack.Message

	if commandFunc, ok := s.registeredCommands[command.Command]; ok != false {
		log.Debugln(command.Command, " found")
		payload = commandFunc(command, ctx)
	} else {
		payload.Msg = slack.Msg{Text: fmt.Sprintf("Unknown command: %s %s", command.Command, command.Text)}
	}
	return payload

}

func (s *SlackBot) FireInteractiveCallback(interactionCallback slack.InteractionCallback, ctx *Context) slack.Message {

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
			callbackFunc(interactionCallback, ctx)
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

func (s *SlackBot) FireCallbackEvent(eventsAPIEvent slackevents.EventsAPIEvent, ctx *Context) {

	innerEvent := eventsAPIEvent.InnerEvent

	eventType := innerEvent.Type

	if eventFunc, ok := s.registeredEvents[eventType]; ok != false {
		eventFunc(eventsAPIEvent, ctx)
	} else {
		log.Debugf("Event %s not registered", eventType)
	}

}
