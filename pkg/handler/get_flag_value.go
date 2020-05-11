package handler

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"
)

func getFlagValue(c *cli.Context, flag Flag, vars map[string]interface{}) (interface{}, error) {
	if c.IsSet(flag.Name) {
		var val interface{}
		switch flag.Type {
		case boolFlagType:
			val = c.Bool(flag.Name)
		default:
			s := c.String(flag.Name)
			if err := validateValueWithValidates(s, flag.Validate); err != nil {
				return nil, fmt.Errorf(flag.Name+" is invalid: %w", err)
			}
			val = s
		}
		return val, nil
	}

	if p := createPrompt(flag.Prompt); p != nil {
		val, err := getValueByPrompt(p, flag.Prompt.Type)
		if err == nil {
			if s, ok := val.(string); ok {
				if err := validateValueWithValidates(s, flag.Validate); err != nil {
					return nil, fmt.Errorf(flag.Name+" is invalid: %w", err)
				}
			}
			return val, nil
		}
	}

	switch flag.Type {
	case boolFlagType:
		// don't ues c.Generic if flag.Type == "bool"
		// the value in the template is treated as false
		return c.Bool(flag.Name), nil
	default:
		if v := c.String(flag.Name); v != "" {
			if err := validateValueWithValidates(v, flag.Validate); err != nil {
				return nil, fmt.Errorf(flag.Name+" is invalid: %w", err)
			}
			return v, nil
		}
		if flag.Required {
			return nil, errors.New(`the flag "` + flag.Name + `" is required`)
		}
		return "", nil
	}
}

func setFlagValues(c *cli.Context, flags []Flag, vars map[string]interface{}) error {
	for _, flag := range flags {
		val, err := getFlagValue(c, flag, vars)
		if err != nil {
			return err
		}
		vars[flag.Name] = val
	}
	return nil
}
