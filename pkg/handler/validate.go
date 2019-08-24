package handler

import (
	"fmt"

	"github.com/pkg/errors"
)

type (
	hasIsSet interface {
		IsSet(string) bool
	}
)

func validateUniqueName(name string, names map[string]struct{}) bool {
	if _, ok := names[name]; ok {
		return false
	}
	names[name] = struct{}{}
	return true
}

func validateFlag(taskName string, flag Flag, flagNames, flagShortNames map[string]struct{}) error {
	if flag.Name == "" {
		return errors.New("the flag name is required: task: " + taskName)
	}
	if len(flag.Short) > 1 {
		return fmt.Errorf(
			"The length of task.short should be 0 or 1. task: %s, flag: %s, short: %s",
			taskName, flag.Name, flag.Short)
	}

	if !validateUniqueName(flag.Name, flagNames) {
		return fmt.Errorf(
			`the flag name duplicates: task: "%s", flag: "%s"`,
			taskName, flag.Name)
	}

	if flag.Short != "" {
		if !validateUniqueName(flag.Short, flagShortNames) {
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
			"The flag type should be either '' or 'string' or 'bool'. task: %s, flag: %s, flag.type: %s",
			taskName, flag.Name, flag.Type)
	}
	return nil
}

func validateArg(taskName string, arg Arg, argNames map[string]struct{}) error {
	if arg.Name == "" {
		return errors.New("the positional argument name is required: task: " + taskName)
	}
	if !validateUniqueName(arg.Name, argNames) {
		return fmt.Errorf(
			`the positional argument name duplicates: task: "%s", arg: "%s"`,
			taskName, arg.Name)
	}
	return nil
}

func validateTask(task Task) error {
	if task.Name == "" {
		return errors.New("the task name is required")
	}
	flagNames := make(map[string]struct{}, len(task.Flags))
	flagShortNames := make(map[string]struct{}, len(task.Flags))
	for _, flag := range task.Flags {
		if err := validateFlag(task.Name, flag, flagNames, flagShortNames); err != nil {
			return err
		}
	}
	argNames := make(map[string]struct{}, len(task.Args))
	for _, arg := range task.Args {
		if err := validateArg(task.Name, arg, argNames); err != nil {
			return err
		}
	}
	return nil
}

func validateConfig(cfg *Config) error {
	taskNames := make(map[string]struct{}, len(cfg.Tasks))
	taskShortNames := make(map[string]struct{}, len(cfg.Tasks))
	for _, task := range cfg.Tasks {
		if !validateUniqueName(task.Name, taskNames) {
			return errors.New(`the task name duplicates: "` + task.Name + `"`)
		}

		if task.Short != "" {
			if !validateUniqueName(task.Short, taskShortNames) {
				return errors.New(`the task short name duplicates: "` + task.Short + `"`)
			}
		}
		if err := validateTask(task); err != nil {
			return err
		}
	}
	return nil
}

func validateFlagRequired(c hasIsSet, flags []Flag) error {
	for _, flag := range flags {
		if !flag.Required {
			continue
		}
		if c.IsSet(flag.Name) {
			continue
		}
		return errors.New(`the flag "` + flag.Name + `" is required`)
	}
	return nil
}
