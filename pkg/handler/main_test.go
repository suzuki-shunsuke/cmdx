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
				Name:      "foo",
				Short:     "f",
				Usage:     "usage",
				InputEnvs: []string{"FOO"},
				Type:      "bool",
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
				Name:      "foo",
				Usage:     "usage",
				Default:   "default value",
				InputEnvs: []string{"FOO"},
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
				Name:      "foo",
				Usage:     "usage",
				InputEnvs: []string{"FOO"},
				Required:  true,
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
		{
			title: "args",
			task: Task{
				Name:        "test",
				Short:       "t",
				Usage:       "usage",
				Description: "description",
				Args: []Arg{
					{
						Name:  "foo",
						Usage: "usage",
					},
				},
			},
			exp: cli.Command{
				Name:        "test",
				ShortName:   "t",
				Usage:       "usage",
				Description: "description",
				Flags:       []cli.Flag{},
				CustomHelpTemplate: `NAME:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}<foo>{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .VisibleFlags}}

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
ARGUMENTS:
   foo  usage`,
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
			assert.Equal(t, d.exp.CustomHelpTemplate, cmd.CustomHelpTemplate)
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

func Test_updateVarsByArgs(t *testing.T) {
	data := []struct {
		title   string
		args    []Arg
		cArgs   []string
		vars    map[string]interface{}
		isErr   bool
		expVars map[string]interface{}
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
		},
		{
			title: "normal",
			args: []Arg{
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
			err := updateVarsByArgs(d.args, d.cArgs, d.vars)
			if err != nil {
				if d.isErr {
					return
				}
				assert.NotNil(t, err)
				return
			}
			assert.False(t, d.isErr)
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
		title string
		task  *Task
		cfg   *Config
		isErr bool
		exp   *Task
	}{
		{
			title: "flags and args are empty",
			task:  &Task{},
			cfg:   &Config{},
			exp: &Task{
				Timeout: Timeout{
					Duration: defaultTimeout,
				},
			},
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
				Environment: map[string]string{
					"ZOO": "zoo",
					"BAR": "bar",
				},
			},
			cfg: &Config{
				InputEnvs: []string{"{{.name}}"},
				Environment: map[string]string{
					"FOO": "foo",
					"BAR": "hello",
				},
			},
			exp: &Task{
				Timeout: Timeout{
					Duration: defaultTimeout,
				},
				Flags: []Flag{
					{
						Name:       "foo",
						InputEnvs:  []string{"FOO"},
						ScriptEnvs: []string{},
					},
				},
				Args: []Arg{
					{
						Name:       "bar",
						InputEnvs:  []string{"BAR"},
						ScriptEnvs: []string{},
					},
				},
				Environment: map[string]string{
					"FOO": "foo",
					"BAR": "bar",
					"ZOO": "zoo",
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := setupTask(d.task, d.cfg)
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

func Test_setupConfig(t *testing.T) {
	data := []struct {
		title string
		cfg   *Config
		exp   *Config
		isErr bool
	}{
		{
			title: "normal",
			cfg: &Config{
				Tasks: []Task{
					{
						Name:   "foo",
						Script: "env",
					},
				},
			},
			exp: &Config{
				Tasks: []Task{
					{
						Name:   "foo",
						Script: "env",
						Timeout: Timeout{
							Duration: defaultTimeout,
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := setupConfig(d.cfg)
			if err != nil {
				if d.isErr {
					return
				}
				assert.NotNil(t, err)
				return
			}
			assert.False(t, d.isErr)
			assert.Equal(t, d.exp, d.cfg)
		})
	}
}

func Test_updateAppWithConfig(t *testing.T) {
	data := []struct {
		title string
		cfg   *Config
	}{
		{
			title: "normal",
			cfg: &Config{
				Tasks: []Task{
					{
						Name:   "foo",
						Script: "env",
					},
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			app := cli.NewApp()
			updateAppWithConfig(app, d.cfg, "/tmp")
		})
	}
}
