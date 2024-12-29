package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

func Test_updateVarsByArgs(t *testing.T) {
	data := []struct {
		title   string
		args    []domain.Arg
		cArgs   []string
		vars    map[string]any
		isErr   bool
		expVars map[string]any
	}{
		{
			title: "args and cArgs is empty",
			expVars: map[string]any{
				"_builtin": map[string]any{
					"args":            []string{},
					"args_string":     "",
					"all_args":        []string{},
					"all_args_string": "",
				},
			},
		},
		{
			title: "normal",
			args: []domain.Arg{
				{
					Name:       "foo",
					ScriptEnvs: []string{"FOO"},
				},
				{
					Name:       "bar",
					ScriptEnvs: []string{"BAR"},
					Default:    "bar-value",
				},
			},
			cArgs: []string{
				"foo-value",
			},
			expVars: map[string]any{
				"foo": "foo-value",
				"bar": "bar-value",
				"_builtin": map[string]any{
					"args":            []string{},
					"args_string":     "",
					"all_args":        []string{"foo-value"},
					"all_args_string": "foo-value",
				},
			},
		},
		{
			title: "required",
			args: []domain.Arg{
				{
					Name:     "foo",
					Required: true,
				},
			},
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			if d.vars == nil {
				d.vars = map[string]any{}
			}
			if d.args == nil {
				d.args = []domain.Arg{}
			}
			if d.cArgs == nil {
				d.cArgs = []string{}
			}
			err := updateVarsByArgs(d.args, d.cArgs, d.vars)
			if err != nil {
				if d.isErr {
					return
				}
				assert.Error(t, err)
				return
			}
			assert.False(t, d.isErr)
			assert.Equal(t, d.expVars, d.vars)
		})
	}
}
