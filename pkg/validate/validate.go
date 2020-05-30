package validate

import (
	"fmt"

	"errors"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
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

func Config(cfg *domain.Config) error {
	taskNames := make(map[string]struct{}, len(cfg.Tasks))
	taskShortNames := make(map[string]struct{}, len(cfg.Tasks))
	for _, task := range cfg.Tasks {
		if !vUniqueName(task.Name, taskNames) {
			return errors.New(`the task name duplicates: "` + task.Name + `"`)
		}

		if task.Short != "" {
			if !vUniqueName(task.Short, taskShortNames) {
				return errors.New(`the task short name duplicates: "` + task.Short + `"`)
			}
		}
		if err := vTask(task); err != nil {
			return err
		}
	}
	return nil
}

func vUniqueName(name string, names map[string]struct{}) bool {
	if _, ok := names[name]; ok {
		return false
	}
	names[name] = struct{}{}
	return true
}

func vFlag(taskName string, flag domain.Flag, flagNames, flagShortNames map[string]struct{}) error {
	if flag.Name == "" {
		return errors.New("the flag name is required: task: " + taskName)
	}
	if len(flag.Short) > 1 {
		return fmt.Errorf(
			"the length of task.short should be 0 or 1. task: %s, flag: %s, short: %s",
			taskName, flag.Name, flag.Short)
	}

	if !vUniqueName(flag.Name, flagNames) {
		return fmt.Errorf(
			`the flag name duplicates: task: "%s", flag: "%s"`,
			taskName, flag.Name)
	}

	if flag.Short != "" {
		if !vUniqueName(flag.Short, flagShortNames) {
			return fmt.Errorf(
				`the flag short name duplicates: task: "%s", flag.short: "%s"`,
				taskName, flag.Short)
		}
	}

	switch flag.Type {
	case "":
	case "bool":
	case "string":
	default:
		return fmt.Errorf(
			"the flag type should be either '' or 'string' or 'bool'. task: %s, flag: %s, flag.type: %s",
			taskName, flag.Name, flag.Type)
	}

	if flag.Prompt.Type != "" {
		if _, ok := flagTypes[flag.Prompt.Type]; !ok {
			return fmt.Errorf(
				"the flag prompt type is invalid: task: %s, flag: %s, prompt: %s",
				taskName, flag.Name, flag.Prompt.Type)
		}
	}

	return nil
}

func vArg(taskName string, arg domain.Arg, argNames map[string]struct{}) error {
	if arg.Name == "" {
		return errors.New("the positional argument name is required: task: " + taskName)
	}
	if !vUniqueName(arg.Name, argNames) {
		return fmt.Errorf(
			`the positional argument name duplicates: task: "%s", arg: "%s"`,
			taskName, arg.Name)
	}
	return nil
}

func vTask(task domain.Task) error {
	if task.Name == "" {
		return errors.New("the task name is required")
	}
	flagNames := make(map[string]struct{}, len(task.Flags))
	flagShortNames := make(map[string]struct{}, len(task.Flags))
	for _, flag := range task.Flags {
		if err := vFlag(task.Name, flag, flagNames, flagShortNames); err != nil {
			return err
		}
	}
	argNames := make(map[string]struct{}, len(task.Args))
	for _, arg := range task.Args {
		if err := vArg(task.Name, arg, argNames); err != nil {
			return err
		}
	}
	if len(task.Tasks) != 0 {
		if task.Script != "" {
			return errors.New("the task `" + task.Name + "` is invalid. when sub tasks are set, 'script' can't be set")
		}
	}
	for _, t := range task.Tasks {
		if err := vTask(t); err != nil {
			return err
		}
	}
	return nil
}
