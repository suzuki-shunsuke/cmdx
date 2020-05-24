package handler

import (
	"github.com/AlecAivazis/survey/v2"
)

var (
	flagTypes = map[string]struct{}{ //nolint:gochecknoglobals
		"input":        {},
		"multiline":    {},
		"password":     {},
		"confirm":      {},
		"select":       {},
		"multi_select": {},
		"editor":       {},
	}
)

func createPrompt(prompt Prompt) survey.Prompt {
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

func getValueByPrompt(prompt survey.Prompt, typ string) (interface{}, error) {
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
