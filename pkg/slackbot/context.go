package slackbot

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"net/http"
)

const (
	SLACK_CONTEXT_HTTP   = 1
	SLACK_CONTEXT_SOCKET = 2
)

type Context struct {
	Type               int
	Api                *slack.Client
	isFinished         bool
	HTTPRequest        *http.Request
	HTTPResponseWriter *http.ResponseWriter
	Socket             *socketmode.Client
	Event              *socketmode.Event
}

func (c Context) IsHTTP() bool {
	return c.Type == SLACK_CONTEXT_HTTP
}
func (c Context) IsSocket() bool {
	return c.Type == SLACK_CONTEXT_SOCKET
}

func (c Context) IsFinished() bool {
	return c.isFinished
}
func (c *Context) Finish() {
	c.isFinished = true
}

func (c *Context) Ack(req socketmode.Request, payload ...interface{}) {
	c.isFinished = true
	c.Socket.Ack(req, payload...)
}

func (s *SlackBot) newHTTPContext(w *http.ResponseWriter, r *http.Request) (context *Context) {
	context = &Context{}
	context.Type = SLACK_CONTEXT_HTTP
	context.Api = s.api
	context.HTTPRequest = r
	context.HTTPResponseWriter = w
	return
}

func (s *SlackBot) newSocketContext(event *socketmode.Event) (context *Context) {
	context = &Context{}
	context.Type = SLACK_CONTEXT_SOCKET
	context.Api = s.api
	context.Socket = s.socket
	context.Event = event
	return
}
