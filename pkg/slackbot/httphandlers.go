package slackbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/humsie/log"
	"github.com/slack-go/slack"
	"io/ioutil"
	"net/http"
)

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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	// restore content back into r.Body
	r.Body = ioutil.NopCloser(bytes.NewReader(body))

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

	s.FireInteractiveCallback(payload)

	w.WriteHeader(http.StatusInternalServerError)

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

	payload := s.FireCommand(command)

	s.renderJSON(w, r, payload)

}

func (s *SlackBot) EventsHandler(w http.ResponseWriter, r *http.Request) {

	if err := s.VerifySignature(w, r); err != nil {
		log.Errorf("Fail to verify SigningSecret: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var payload map[string]interface{}

	jsondec := json.NewDecoder(r.Body)
	err := jsondec.Decode(&payload)
	if err != nil {
		fmt.Printf("Could not parse action response JSON: %v", err)
	}

	switch payload["type"] {
	case "url_verification":
		w.WriteHeader(200)
		w.Write([]byte(payload["challenge"].(string)))
		return
	default:
		log.Debugln("Got a event of type: ", payload["type"].(string))
	}

	w.WriteHeader(401)

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

func (s *SlackBot) renderError(w http.ResponseWriter, r *http.Request, err error) {
	b := []byte(fmt.Sprintf("There was an error: %s", err))

	debug := true

	if debug == false {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.Write(b)
	}

}
