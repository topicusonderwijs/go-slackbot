package slackbot

import (
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"net/http"
)

type Context interface {
	IsFinished() bool
	Finish()
}

type BaseContext struct {
	Api        *slack.Client
	isFinished bool
}

func (c BaseContext) IsFinished() bool {
	return c.isFinished
}
func (c *BaseContext) Finish() {
	c.isFinished = true
}

type HTTPContext struct {
	BaseContext
	req  *http.Request
	resp *http.ResponseWriter
}

type SocketContext struct {
	BaseContext
	Socket *socketmode.Client
	Event  *socketmode.Event
}

func (c *SocketContext) Ack(req socketmode.Request, payload ...interface{}) {
	c.isFinished = true
	c.Socket.Ack(req, payload...)
}

func (s *SlackBot) newHTTPContext(w *http.ResponseWriter, r *http.Request) (context *HTTPContext) {
	context = &HTTPContext{}
	context.Api = s.api
	context.req = r
	context.resp = w
	return
}

func (s *SlackBot) newSocketContext(event *socketmode.Event) (context *SocketContext) {
	context = &SocketContext{}
	context.Api = s.api
	context.Socket = s.socket
	context.Event = event
	return
}
