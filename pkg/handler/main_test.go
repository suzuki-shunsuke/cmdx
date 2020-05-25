package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

func Test_setupApp(t *testing.T) {
	app := cli.NewApp()
	setupApp(app)
	assert.Equal(t, "cmdx", app.Name)
	assert.Equal(t, appUsage, app.Usage)
	assert.Equal(t, domain.Version, app.Version)
	assert.Equal(t, []*cli.Author{
		{
			Name: "Shunsuke Suzuki",
		},
	}, app.Authors)
}

func Test_newFlag(t *testing.T) {
	data := []struct {
		title string
		flag  domain.Flag
		exp   cli.Flag
	}{
		{
			title: "bool",
			flag: domain.Flag{
				Name:      "foo",
				Short:     "f",
				Usage:     "usage",
				InputEnvs: []string{"FOO"},
				Type:      "bool",
			},
			exp: &cli.BoolFlag{
				Name:    "foo, f",
				Usage:   "usage",
				EnvVars: []string{"FOO"},
			},
		},
		{
			title: "string",
			flag: domain.Flag{
				Name:      "foo",
				Usage:     "usage",
				Default:   "default value",
				InputEnvs: []string{"FOO"},
			},
			exp: &cli.StringFlag{
				Name:    "foo",
				Usage:   "usage",
				Value:   "default value",
				EnvVars: []string{"FOO"},
			},
		},
		{
			title: "required",
			flag: domain.Flag{
				Name:      "foo",
				Usage:     "usage",
				InputEnvs: []string{"FOO"},
				Required:  true,
			},
			exp: &cli.StringFlag{
				Name:    "foo",
				Usage:   "usage",
				EnvVars: []string{"FOO"},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			assert.Equal(t, d.exp, newFlag(d.flag))
		})
	}
}

func Test_convertTaskToCommand(t *testing.T) {
	data := []struct {
		title string
		task  domain.Task
		exp   cli.Command
	}{
		{
			title: "no flag",
			task: domain.Task{
				Name:        "test",
				Short:       "t",
				Usage:       "usage",
				Description: "description",
			},
			exp: cli.Command{
				Name:               "test",
				Aliases:            []string{"t"},
				Usage:              "usage",
				Description:        "description",
				CustomHelpTemplate: cli.CommandHelpTemplate,
				Flags:              []cli.Flag{},
			},
		},
		{
			title: "flag",
			task: domain.Task{
				Name:        "test",
				Short:       "t",
				Usage:       "usage",
				Description: "description",
				Flags: []domain.Flag{
					{
						Name: "foo",
					},
				},
			},
			exp: cli.Command{
				Name:               "test",
				Aliases:            []string{"t"},
				Usage:              "usage",
				Description:        "description",
				CustomHelpTemplate: cli.CommandHelpTemplate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "foo",
					},
				},
			},
		},
		{
			title: "args",
			task: domain.Task{
				Name:        "test",
				Short:       "t",
				Usage:       "usage",
				Description: "description",
				Args: []domain.Arg{
					{
						Name:  "foo",
						Usage: "usage",
					},
				},
			},
			exp: cli.Command{
				Name:        "test",
				Aliases:     []string{"t"},
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
		d := d
		t.Run(d.title, func(t *testing.T) {
			cmd := convertTaskToCommand(d.task, &GlobalFlags{})
			assert.Equal(t, d.exp.Name, cmd.Name)
			assert.Equal(t, d.exp.Aliases, cmd.Aliases)
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
		d := d
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
		args    []domain.Arg
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
		d := d
		t.Run(d.title, func(t *testing.T) {
			if d.vars == nil {
				d.vars = map[string]interface{}{}
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
		task  *domain.Task
		cfg   *domain.Config
		isErr bool
		exp   *domain.Task
	}{
		{
			title: "flags and args are empty",
			task:  &domain.Task{},
			cfg:   &domain.Config{},
			exp: &domain.Task{
				Timeout: domain.Timeout{
					Duration: defaultTimeout,
				},
				Environment: map[string]string{},
			},
		},
		{
			title: "set environment variable",
			task:  &domain.Task{},
			cfg: &domain.Config{
				Environment: map[string]string{
					"FOO": "foo",
				},
			},
			exp: &domain.Task{
				Timeout: domain.Timeout{
					Duration: defaultTimeout,
				},
				Environment: map[string]string{
					"FOO": "foo",
				},
			},
		},
		{
			title: "normal",
			task: &domain.Task{
				Flags: []domain.Flag{
					{
						Name: "foo",
					},
				},
				Args: []domain.Arg{
					{
						Name: "bar",
					},
				},
				Environment: map[string]string{
					"ZOO": "zoo",
					"BAR": "bar",
				},
			},
			cfg: &domain.Config{
				InputEnvs: []string{"{{.name}}"},
				Environment: map[string]string{
					"FOO": "foo",
					"BAR": "hello",
				},
			},
			exp: &domain.Task{
				Timeout: domain.Timeout{
					Duration: defaultTimeout,
				},
				Flags: []domain.Flag{
					{
						Name:       "foo",
						InputEnvs:  []string{"FOO"},
						ScriptEnvs: []string{},
					},
				},
				Args: []domain.Arg{
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
		d := d
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
		cfg   *domain.Config
		exp   *domain.Config
		isErr bool
	}{
		{
			title: "normal",
			cfg: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   "foo",
						Script: "env",
					},
				},
			},
			exp: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   "foo",
						Script: "env",
						Timeout: domain.Timeout{
							Duration: defaultTimeout,
						},
						Environment: map[string]string{},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
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
		cfg   *domain.Config
	}{
		{
			title: "normal",
			cfg: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   "foo",
						Script: "env",
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			app := cli.NewApp()
			updateAppWithConfig(app, d.cfg, &GlobalFlags{WorkingDir: "/tmp"})
		})
	}
}

func Test_getHelp(t *testing.T) {
	data := []struct {
		title string
		txt   string
		task  domain.Task
		exp   string
	}{
		{
			title: "not update",
			txt:   "hello",
			exp:   "hello",
		},
	}
	for _, d := range data {
		assert.Equal(t, d.exp, getHelp(d.txt, d.task))
	}
}
