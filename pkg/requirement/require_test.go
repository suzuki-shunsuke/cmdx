package requirement

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecker_Exec(t *testing.T) {
	data := []struct {
		title string
		execs []string
		isErr bool
	}{
		{
			title: "no validation",
			execs: nil,
			isErr: false,
		},
	}
	checker := New()
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := checker.Exec(d.execs)
			if d.isErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestChecker_Env(t *testing.T) {
	data := []struct {
		title   string
		envs    []string
		setEnvs map[string]string
		isErr   bool
	}{
		{
			title: "no validation",
			envs:  nil,
			isErr: false,
		},
		{
			title: "ok",
			envs:  []string{"BAR", "ZOO"},
			setEnvs: map[string]string{
				"FOO": "foo",
				"BAR": "bar",
				"ZOO": "zoo",
			},
			isErr: false,
		},
		{
			title:   "FOO is required",
			envs:    []string{"FOO"},
			setEnvs: nil,
			isErr:   true,
		},
		{
			title:   "FOO or BAR is required",
			envs:    []string{"FOO", "BAR"},
			setEnvs: nil,
			isErr:   true,
		},
	}
	defer os.Clearenv()
	checker := New()
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			os.Clearenv()
			for k, v := range d.setEnvs {
				t.Setenv(k, v)
			}
			err := checker.Env(d.envs)
			if d.isErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
