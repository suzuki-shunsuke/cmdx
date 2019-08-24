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

func createPrompt(flagName string, prompt Prompt) survey.Prompt {
	switch prompt.Type {
	case "":
		return nil
	case "input":
		return &survey.Input{Message: flagName}
	case "multiline":
		return &survey.Multiline{Message: flagName}
	case "password":
		return &survey.Password{Message: flagName}
	case "confirm":
		return &survey.Confirm{Message: flagName}
	case "select":
		return &survey.Select{Message: flagName, Options: prompt.Options}
	case "multi_select":
		return &survey.MultiSelect{Message: flagName, Options: prompt.Options}
	case "editor":
		return &survey.Editor{Message: flagName}
	}
	return nil
}
