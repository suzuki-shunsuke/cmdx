package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_bindScriptEnvs(t *testing.T) {
	data := []struct {
		title      string
		envs       []string
		vars       map[string]interface{}
		scriptEnvs map[string][]string
		exp        []string
	}{
		{
			title: "nil",
		},
		{
			title: "nil",
			envs:  []string{"FOO=foo"},
			vars: map[string]interface{}{
				"man":  true,
				"age":  "10",
				"list": []string{"foo", "bar"},
			},
			scriptEnvs: map[string][]string{
				"age":  {"AGE", "ZOO"},
				"list": {"BAR"},
				"man":  {"BOO"},
			},
			exp: []string{
				"FOO=foo",
				"AGE=10", "ZOO=10",
				"BAR=foo,bar",
				"BOO=true",
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			assert.ElementsMatch(t, d.exp, bindScriptEnvs(d.envs, d.vars, d.scriptEnvs))
		})
	}
}
