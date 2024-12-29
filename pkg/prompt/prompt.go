package prompt

import (
	"github.com/AlecAivazis/survey/v2"
)

const (
	confirmPromptType = "confirm"
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
	case "input":
		return &survey.Input{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case "multiline":
		return &survey.Multiline{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case "password":
		return &survey.Password{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case "confirm":
		return &survey.Confirm{
			Message: prompt.Message,
			Help:    prompt.Help,
		}
	case "select":
		return &survey.Select{
			Message: prompt.Message,
			Help:    prompt.Help,
			Options: prompt.Options,
		}
	case "multi_select":
		return &survey.MultiSelect{
			Message: prompt.Message,
			Help:    prompt.Help,
			Options: prompt.Options,
		}
	case "editor":
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
	case "select":
		ans := ""
		if err := survey.AskOne(prompt, &ans); err != nil {
			return nil, err
		}
		return ans, nil
	case "multi_select":
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
