package flag

import (
	"errors"
	"fmt"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	"github.com/suzuki-shunsuke/cmdx/pkg/prompt"
	"github.com/suzuki-shunsuke/cmdx/pkg/validate"
	"github.com/urfave/cli/v2"
)

const (
	boolFlagType = "bool"
)

func getFlagValue(c *cli.Context, flag domain.Flag) (any, error) {
	if c.IsSet(flag.Name) {
		switch flag.Type {
		case boolFlagType:
			return c.Bool(flag.Name), nil
		default:
			s := c.String(flag.Name)
			if err := validate.ValueWithValidates(s, flag.Validate); err != nil {
				return nil, fmt.Errorf("%s is invalid: %w", flag.Name, err)
			}
			return s, nil
		}
	}

	if p := prompt.Create(flag.Prompt); p != nil {
		val, err := prompt.GetValue(p, flag.Prompt.Type)
		if err == nil {
			if s, ok := val.(string); ok {
				if err := validate.ValueWithValidates(s, flag.Validate); err != nil {
					return nil, fmt.Errorf("%s is invalid: %w", flag.Name, err)
				}
			}
			return val, nil
		}
	}

	switch flag.Type {
	case boolFlagType:
		// don't use c.Generic if flag.Type == "bool"
		// the value in the template is treated as false
		return c.Bool(flag.Name), nil
	default:
		if v := c.String(flag.Name); v != "" {
			if err := validate.ValueWithValidates(v, flag.Validate); err != nil {
				return nil, fmt.Errorf("%s is invalid: %w", flag.Name, err)
			}
			return v, nil
		}
		if flag.Required {
			return nil, errors.New(`the flag "` + flag.Name + `" is required`)
		}
		return "", nil
	}
}

func SetValues(c *cli.Context, flags []domain.Flag, vars map[string]any) error {
	for _, flag := range flags {
		val, err := getFlagValue(c, flag)
		if err != nil {
			return err
		}
		vars[flag.Name] = val
	}
	return nil
}
