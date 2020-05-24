package requirement

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

type Checker struct{}

func New() *Checker {
	return &Checker{}
}

func (*Checker) Exec(requires []string) error {
	if len(requires) == 0 {
		return nil
	}
	for _, require := range requires {
		if _, err := exec.LookPath(require); err == nil {
			return nil
		}
	}
	if len(requires) == 1 {
		return errors.New(requires[0] + " is required")
	}
	return errors.New("one of the following is required: " + strings.Join(requires, ", "))
}

func (*Checker) Env(requires []string) error {
	if len(requires) == 0 {
		return nil
	}
	for _, require := range requires {
		if os.Getenv(require) != "" {
			return nil
		}
	}
	if len(requires) == 1 {
		return errors.New("the environment variable '" + requires[0] + "' is required")
	}
	return errors.New("one of the following environment variables is required: " + strings.Join(requires, ", "))
}
