package slackbot

import (
	"testing"

	"github.com/slack-go/slack"
)

// TestFireInteractiveCallbackBlockActionsRouting covers the three block_actions
// dispatch cases: routing by Value (buttons), the SelectedOption.Value fallback
// (per-value select handlers), and the additive action_id dispatch (a single
// handler for a whole select).
func TestFireInteractiveCallbackBlockActionsRouting(t *testing.T) {
	tests := []struct {
		name         string
		register     []string
		action       slack.BlockAction
		wantFired    string
		wantSelected string
	}{
		{
			name:      "button routes by Value",
			register:  []string{"button_pressed"},
			action:    slack.BlockAction{ActionID: "some_action", Value: "button_pressed"},
			wantFired: "button_pressed",
		},
		{
			name:     "select with per-value handler routes by SelectedOption.Value",
			register: []string{"option_a"},
			action: slack.BlockAction{
				ActionID:       "businessline",
				SelectedOption: slack.OptionBlockObject{Value: "option_a"},
			},
			wantFired:    "option_a",
			wantSelected: "option_a",
		},
		{
			name:     "select with action_id handler routes by ActionID",
			register: []string{"businessline"},
			action: slack.BlockAction{
				ActionID:       "businessline",
				SelectedOption: slack.OptionBlockObject{Value: "option_a"},
			},
			wantFired:    "businessline",
			wantSelected: "option_a",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bot := &SlackBot{}
			bot.Setup()

			var fired, selected string
			for _, id := range tc.register {
				id := id
				err := bot.RegisterInteractionCallback(
					slack.InteractionTypeBlockActions,
					id,
					func(cb slack.InteractionCallback, ctx *Context) slack.Message {
						fired = id
						if len(cb.ActionCallback.BlockActions) > 0 {
							selected = cb.ActionCallback.BlockActions[0].SelectedOption.Value
						}
						return slack.Message{}
					},
				)
				if err != nil {
					t.Fatalf("register %s: %v", id, err)
				}
			}

			cb := slack.InteractionCallback{Type: slack.InteractionTypeBlockActions}
			cb.ActionCallback.BlockActions = []*slack.BlockAction{&tc.action}

			bot.FireInteractiveCallback(cb, &Context{})

			if fired != tc.wantFired {
				t.Errorf("fired handler = %q, want %q", fired, tc.wantFired)
			}
			if tc.wantSelected != "" && selected != tc.wantSelected {
				t.Errorf("selected option = %q, want %q", selected, tc.wantSelected)
			}
		})
	}
}
