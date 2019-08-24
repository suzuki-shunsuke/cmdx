package handler

import (
	"github.com/AlecAivazis/survey/v2"
)

var (
	flagTypes = map[string]struct{}{
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
