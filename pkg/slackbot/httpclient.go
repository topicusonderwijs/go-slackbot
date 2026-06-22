package slackbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io"
	"net/http"
)

func (s *SlackBot) SetHTTPHandleFunctions(http *http.ServeMux) {

	http.HandleFunc("/events", s.DefaultHandler)
	http.HandleFunc("/slack/events", s.EventsHandler)
	http.HandleFunc("/slack/load-options", s.DefaultHandler)

	http.HandleFunc("/slack/actions", s.ActionsHandler)
	http.HandleFunc("/slack/commands", s.CommandsHandler)

}

func (s *SlackBot) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Got request on: %s", r.RequestURI)
	w.WriteHeader(http.StatusNotFound)
	return
}

func (s *SlackBot) VerifySignature(w http.ResponseWriter, r *http.Request) (err error) {

	verifier, err := slack.NewSecretsVerifier(r.Header, s.config.signSecret)
	if err != nil {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	// restore content back into r.Body
	r.Body = io.NopCloser(bytes.NewReader(body))

	if _, err = verifier.Write(body); err != nil {
		return
	}
	if err = verifier.Ensure(); err != nil {
		return
	}

	log.Debugf("Successfully verified SigningSecret")

	return nil

}

func (s *SlackBot) ActionsHandler(w http.ResponseWriter, r *http.Request) {

	if err := s.VerifySignature(w, r); err != nil {
		log.Errorf("Fail to verify SigningSecret: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		fmt.Printf("Could not parse action response JSON: %v", err)
	}

	ctx := s.newHTTPContext(w, r)
	response := s.FireInteractiveCallback(payload, ctx)

	if !ctx.IsFinished() {
		s.renderJSON(w, r, response)
	}

}

func (s *SlackBot) CommandsHandler(w http.ResponseWriter, r *http.Request) {

	if err := s.VerifySignature(w, r); err != nil {
		log.Errorf("Fail to verify SigningSecret: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	command, err := slack.SlashCommandParse(r)

	log.Debugf("Got responseUrl: %s", command.ResponseURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx := s.newHTTPContext(w, r)
	payload := s.FireSlashCommand(command, ctx)

	if !ctx.IsFinished() {
		s.renderJSON(w, r, payload)
	}

}

func (s *SlackBot) EventsHandler(w http.ResponseWriter, r *http.Request) {

	if err := s.VerifySignature(w, r); err != nil {
		log.Errorf("Fail to verify SigningSecret: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Could not read event body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// The signature is already verified in VerifySignature.
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		log.Errorf("Could not parse event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			log.Errorf("Could not parse challenge: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(res.Challenge))
	case slackevents.CallbackEvent:
		ctx := s.newHTTPContext(w, r)
		s.FireCallbackEvent(eventsAPIEvent, ctx)
		w.WriteHeader(http.StatusOK)
	default:
		log.Debugln("Unhandled event type: ", eventsAPIEvent.Type)
		w.WriteHeader(http.StatusOK)
	}

}

func (s *SlackBot) renderJSON(w http.ResponseWriter, r *http.Request, object interface{}) {

	accept := r.Header.Get("Accept")
	if accept != "*/*" {
		log.Debugf("Accept is %s", accept)
	}

	b, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

}
