package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

func Test_setupApp(t *testing.T) {
	app := cli.NewApp()
	setupApp(app)
	assert.Equal(t, "cmdx", app.Name)
	assert.Equal(t, appUsage, app.Usage)
	assert.Equal(t, domain.Version, app.Version)
	assert.Equal(t, []cli.Author{
		{
			Name: "Shunsuke Suzuki",
		},
	}, app.Authors)
}

func Test_newFlag(t *testing.T) {
	data := []struct {
		title string
		flag  Flag
		exp   cli.Flag
	}{
		{
			title: "bool",
			flag: Flag{
				Name:     "foo",
				Short:    "f",
				Usage:    "usage",
				BindEnvs: []string{"FOO"},
				Type:     "bool",
			},
			exp: cli.BoolFlag{
				Name:   "foo, f",
				Usage:  "usage",
				EnvVar: "FOO",
			},
		},
		{
			title: "string",
			flag: Flag{
				Name:     "foo",
				Usage:    "usage",
				Default:  "default value",
				BindEnvs: []string{"FOO"},
			},
			exp: cli.StringFlag{
				Name:   "foo",
				Usage:  "usage",
				Value:  "default value",
				EnvVar: "FOO",
			},
		},
		{
			title: "required",
			flag: Flag{
				Name:     "foo",
				Usage:    "usage",
				BindEnvs: []string{"FOO"},
				Required: true,
			},
			exp: cli.StringFlag{
				Name:   "foo",
				Usage:  "usage",
				EnvVar: "FOO",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			assert.Equal(t, d.exp, newFlag(d.flag))
		})
	}
}

func Test_convertTaskToCommand(t *testing.T) {
	data := []struct {
		title string
		task  Task
		exp   cli.Command
	}{
		{
			title: "no flag",
			task: Task{
				Name:        "test",
				Short:       "t",
				Usage:       "usage",
				Description: "description",
			},
			exp: cli.Command{
				Name:        "test",
				ShortName:   "t",
				Usage:       "usage",
				Description: "description",
				Flags:       []cli.Flag{},
			},
		},
		{
			title: "flag",
			task: Task{
				Name:        "test",
				Short:       "t",
				Usage:       "usage",
				Description: "description",
				Flags: []Flag{
					{
						Name: "foo",
					},
				},
			},
			exp: cli.Command{
				Name:        "test",
				ShortName:   "t",
				Usage:       "usage",
				Description: "description",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name: "foo",
					},
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			cmd := convertTaskToCommand(d.task, "")
			assert.Equal(t, d.exp.Name, cmd.Name)
			assert.Equal(t, d.exp.ShortName, cmd.ShortName)
			assert.Equal(t, d.exp.Usage, cmd.Usage)
			assert.Equal(t, d.exp.Flags, cmd.Flags)
			assert.Equal(t, d.exp.Description, cmd.Description)
		})
	}
}

func Test_renderTemplate(t *testing.T) {
	data := []struct {
		title string
		base  string
		data  interface{}
		isErr bool
		exp   string
	}{
		{
			title: "normal",
			base:  "foo {{.source}}",
			data:  map[string]string{"source": "bar"},
			exp:   "foo bar",
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			s, err := renderTemplate(d.base, d.data)
			if err != nil {
				if d.isErr {
					return
				}
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, d.exp, s)
		})
	}
}

func Test_updateVarsAndEnvsByArgs(t *testing.T) {
	data := []struct {
		title   string
		args    []Arg
		cArgs   []string
		vars    map[string]interface{}
		isErr   bool
		expVars map[string]interface{}
		expEnvs []string
	}{
		{
			title: "args and cArgs is empty",
			expVars: map[string]interface{}{
				"_builtin": map[string]interface{}{
					"args":            []string{},
					"args_string":     "",
					"all_args":        []string{},
					"all_args_string": "",
				},
			},
			expEnvs: []string{},
		},
		{
			title: "normal",
			args: []Arg{
				{
					Name:     "foo",
					BindEnvs: []string{"FOO"},
				},
				{
					Name:     "bar",
					BindEnvs: []string{"BAR"},
					Default:  "bar-value",
				},
			},
			cArgs: []string{
				"foo-value",
			},
			expVars: map[string]interface{}{
				"foo": "foo-value",
				"bar": "bar-value",
				"_builtin": map[string]interface{}{
					"args":            []string{},
					"args_string":     "",
					"all_args":        []string{"foo-value"},
					"all_args_string": "foo-value",
				},
			},
			expEnvs: []string{"FOO=foo-value", "BAR=bar-value"},
		},
		{
			title: "required",
			args: []Arg{
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
				d.vars = map[string]interface{}{}
			}
			if d.args == nil {
				d.args = []Arg{}
			}
			if d.cArgs == nil {
				d.cArgs = []string{}
			}
			envs, err := updateVarsAndEnvsByArgs(d.args, d.cArgs, d.vars)
			if err != nil {
				if d.isErr {
					return
				}
				assert.NotNil(t, err)
				return
			}
			assert.False(t, d.isErr)
			assert.Equal(t, d.expEnvs, envs)
			assert.Equal(t, d.expVars, d.vars)
		})
	}
}

func Test_setupEnvs(t *testing.T) {
	envs, err := setupEnvs([]string{"{{.name}}-zoo"}, "foo")
	assert.Nil(t, err)
	assert.Equal(t, []string{"FOO_ZOO"}, envs)
}

func Test_setupTask(t *testing.T) {
	data := []struct {
		title    string
		task     *Task
		bindEnvs []string
		isErr    bool
		exp      *Task
	}{
		{
			title: "flags and args are empty",
			task:  &Task{},
			exp:   &Task{},
		},
		{
			title: "normal",
			task: &Task{
				Flags: []Flag{
					{
						Name: "foo",
					},
				},
				Args: []Arg{
					{
						Name: "bar",
					},
				},
			},
			bindEnvs: []string{"{{.name}}"},
			exp: &Task{
				Flags: []Flag{
					{
						Name:     "foo",
						BindEnvs: []string{"FOO"},
					},
				},
				Args: []Arg{
					{
						Name:     "bar",
						BindEnvs: []string{"BAR"},
					},
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := setupTask(d.task, d.bindEnvs)
			if err != nil {
				if d.isErr {
					return
				}
				assert.NotNil(t, err)
				return
			}
			assert.False(t, d.isErr)
			assert.Equal(t, d.exp, d.task)
		})
	}
}
