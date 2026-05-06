package prompt

import (
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
)

const (
	testMessage = "message"
	testHelp    = "help"
	testBlue    = "blue"
	testGreen   = "green"
)

func Test_createPrompt(t *testing.T) {
	data := []struct {
		title  string
		prompt Prompt
		exp    survey.Prompt
	}{
		{
			title: inputPromptType,
			prompt: Prompt{
				Type:    inputPromptType,
				Message: testMessage,
				Help:    testHelp,
			},
			exp: &survey.Input{
				Message: testMessage,
				Help:    testHelp,
			},
		},
		{
			title: multilinePromptType,
			prompt: Prompt{
				Type:    multilinePromptType,
				Message: testMessage,
				Help:    testHelp,
			},
			exp: &survey.Multiline{
				Message: testMessage,
				Help:    testHelp,
			},
		},
		{
			title: passwordPromptType,
			prompt: Prompt{
				Type:    passwordPromptType,
				Message: testMessage,
				Help:    testHelp,
			},
			exp: &survey.Password{
				Message: testMessage,
				Help:    testHelp,
			},
		},
		{
			title: confirmPromptType,
			prompt: Prompt{
				Type:    confirmPromptType,
				Message: testMessage,
				Help:    testHelp,
			},
			exp: &survey.Confirm{
				Message: testMessage,
				Help:    testHelp,
			},
		},
		{
			title: editorPromptType,
			prompt: Prompt{
				Type:    editorPromptType,
				Message: testMessage,
				Help:    testHelp,
			},
			exp: &survey.Editor{
				Message:       testMessage,
				Help:          testHelp,
				HideDefault:   true,
				AppendDefault: true,
			},
		},
		{
			title: selectPromptType,
			prompt: Prompt{
				Type:    selectPromptType,
				Message: testMessage,
				Help:    testHelp,
				Options: []string{testBlue, testGreen},
			},
			exp: &survey.Select{
				Message: testMessage,
				Help:    testHelp,
				Options: []string{testBlue, testGreen},
			},
		},
		{
			title: multiSelectPromptType,
			prompt: Prompt{
				Type:    multiSelectPromptType,
				Message: testMessage,
				Help:    testHelp,
				Options: []string{testBlue, testGreen},
			},
			exp: &survey.MultiSelect{
				Message: testMessage,
				Help:    testHelp,
				Options: []string{testBlue, testGreen},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			assert.Equal(t, d.exp, Create(d.prompt))
		})
	}
}
