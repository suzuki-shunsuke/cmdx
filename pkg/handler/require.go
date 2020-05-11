package handler

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func requireExec(execs []StrList) error {
	for _, requires := range execs {
		if len(requires) == 0 {
			continue
		}
		f := false
		for _, require := range requires {
			if _, err := exec.LookPath(require); err == nil {
				f = true
				break
			}
		}
		if !f {
			if len(requires) == 1 {
				return errors.New(requires[0] + " is required")
			}
			return errors.New("one of the following is required: " + strings.Join(requires, ", "))
		}
	}
	return nil
}

func requireEnv(envs []StrList) error {
	for _, requires := range envs {
		if len(requires) == 0 {
			continue
		}
		f := false
		for _, require := range requires {
			if os.Getenv(require) != "" {
				f = true
				break
			}
		}
		if !f {
			if len(requires) == 1 {
				return errors.New("the environment variable '" + requires[0] + "' is required")
			}
			return errors.New("one of the following environment variables is required: " + strings.Join(requires, ", "))
		}
	}
	return nil
}
