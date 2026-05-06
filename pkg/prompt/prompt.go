package prompt

import (
	"github.com/AlecAivazis/survey/v2"
)

const (
	inputPromptType       = "input"
	multilinePromptType   = "multiline"
	passwordPromptType    = "password"
	confirmPromptType     = "confirm"
	selectPromptType      = "select"
	multiSelectPromptType = "multi_select"
	editorPromptType      = "editor"
)

type Prompt struct {
	Type    string   `json:"type"`
	Message string   `json:"message,omitempty"`
	Help    string   `json:"help,omitempty"`
	Options []string `json:"options,omitempty"`
}

func Create(prompt Prompt) survey.Prompt {
	switch prompt.Type {
	case "":
		return nil
	case inputPromptType:
		return &survey.Input{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case multilinePromptType:
		return &survey.Multiline{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case passwordPromptType:
		return &survey.Password{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case confirmPromptType:
		return &survey.Confirm{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case selectPromptType:
		return &survey.Select{
			Message: prompt.Message,
			Help:    prompt.Help,
			Options: prompt.Options,
		}
	case multiSelectPromptType:
		return &survey.MultiSelect{
			Message: prompt.Message,
			Help:    prompt.Help,
			Options: prompt.Options,
		}
	case editorPromptType:
		return &survey.Editor{
			Message:       prompt.Message,
			Help:          prompt.Help,
			HideDefault:   true,
			AppendDefault: true,
		}
	}
	return nil
}

func GetValue(prompt survey.Prompt, typ string) (any, error) {
	switch typ {
	case confirmPromptType:
		ans := false
		err := survey.AskOne(prompt, &ans)
		return ans, err
	case selectPromptType:
		ans := ""
		if err := survey.AskOne(prompt, &ans); err != nil {
			return nil, err
		}
		return ans, nil
	case multiSelectPromptType:
		ans := []string{}
		if err := survey.AskOne(prompt, &ans); err != nil {
			return nil, err
		}
		return ans, nil
	default:
		ans := ""
		return ans, survey.AskOne(prompt, &ans)
	}
}
