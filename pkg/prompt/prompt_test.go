package prompt

import (
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
)

func Test_createPrompt(t *testing.T) {
	data := []struct {
		title  string
		prompt Prompt
		exp    survey.Prompt
	}{
		{
			title: "input",
			prompt: Prompt{
				Type:    "input",
				Message: "message",
				Help:    "help",
			},
			exp: &survey.Input{
				Message: "message",
				Help:    "help",
			},
		},
		{
			title: "multiline",
			prompt: Prompt{
				Type:    "multiline",
				Message: "message",
				Help:    "help",
			},
			exp: &survey.Multiline{
				Message: "message",
				Help:    "help",
			},
		},
		{
			title: "password",
			prompt: Prompt{
				Type:    "password",
				Message: "message",
				Help:    "help",
			},
			exp: &survey.Password{
				Message: "message",
				Help:    "help",
			},
		},
		{
			title: "confirm",
			prompt: Prompt{
				Type:    "confirm",
				Message: "message",
				Help:    "help",
			},
			exp: &survey.Confirm{
				Message: "message",
				Help:    "help",
			},
		},
		{
			title: "editor",
			prompt: Prompt{
				Type:    "editor",
				Message: "message",
				Help:    "help",
			},
			exp: &survey.Editor{
				Message:       "message",
				Help:          "help",
				HideDefault:   true,
				AppendDefault: true,
			},
		},
		{
			title: "select",
			prompt: Prompt{
				Type:    "select",
				Message: "message",
				Help:    "help",
				Options: []string{"blue", "green"},
			},
			exp: &survey.Select{
				Message: "message",
				Help:    "help",
				Options: []string{"blue", "green"},
			},
		},
		{
			title: "multi_select",
			prompt: Prompt{
				Type:    "multi_select",
				Message: "message",
				Help:    "help",
				Options: []string{"blue", "green"},
			},
			exp: &survey.MultiSelect{
				Message: "message",
				Help:    "help",
				Options: []string{"blue", "green"},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			assert.Equal(t, d.exp, Create(d.prompt))
		})
	}
}
