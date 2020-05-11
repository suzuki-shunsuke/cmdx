package handler

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"
)

func setFlagValue(c *cli.Context, flag Flag, vars map[string]interface{}) error {
	if c.IsSet(flag.Name) {
		var val interface{}
		switch flag.Type {
		case boolFlagType:
			val = c.Bool(flag.Name)
		default:
			s := c.String(flag.Name)
			if err := validateValueWithValidates(s, flag.Validate); err != nil {
				return fmt.Errorf(flag.Name+" is invalid: %w", err)
			}
			val = s
		}

		vars[flag.Name] = val
		return nil
	}

	if p := createPrompt(flag.Prompt); p != nil {
		val, err := getValueByPrompt(p, flag.Prompt.Type)
		if err == nil {
			if s, ok := val.(string); ok {
				if err := validateValueWithValidates(s, flag.Validate); err != nil {
					return fmt.Errorf(flag.Name+" is invalid: %w", err)
				}
			}

			vars[flag.Name] = val
			return nil
		}
	}

	switch flag.Type {
	case boolFlagType:
		// don't ues c.Generic if flag.Type == "bool"
		// the value in the template is treated as false
		vars[flag.Name] = c.Bool(flag.Name)
	default:
		if v := c.String(flag.Name); v != "" {
			if err := validateValueWithValidates(v, flag.Validate); err != nil {
				return fmt.Errorf(flag.Name+" is invalid: %w", err)
			}

			vars[flag.Name] = v
			return nil
		}
		if flag.Required {
			return errors.New(`the flag "` + flag.Name + `" is required`)
		}
		vars[flag.Name] = ""
	}

	return nil
}

func setFlagValues(c *cli.Context, flags []Flag, vars map[string]interface{}) error {
	for _, flag := range flags {
		if err := setFlagValue(c, flag, vars); err != nil {
			return err
		}
	}
	return nil
}
