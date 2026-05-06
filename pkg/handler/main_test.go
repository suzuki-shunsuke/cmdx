package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	"github.com/urfave/cli/v2"
)

const (
	envFOO         = "FOO"
	envBAR         = "BAR"
	valBar         = "bar"
	valFoo         = "foo"
	valUsage       = "usage"
	valTest        = "test"
	valDescription = "description"
	titleNormal    = "normal"
	valHello       = "hello"
	valEnv         = "env"
)

func Test_setupApp(t *testing.T) {
	app := cli.NewApp()
	flags := &LDFlags{
		Version: "v1.6.0",
	}
	setupApp(app, flags)
	assert.Equal(t, "cmdx", app.Name)
	assert.Equal(t, appUsage, app.Usage)
	assert.Equal(t, flags.AppVersion(), app.Version)
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
				Name:      valFoo,
				Short:     "f",
				Usage:     valUsage,
				InputEnvs: []string{envFOO},
				Type:      "bool",
			},
			exp: &cli.BoolFlag{
				Name:    valFoo,
				Usage:   valUsage,
				EnvVars: []string{envFOO},
				Aliases: []string{"f"},
			},
		},
		{
			title: "string",
			flag: domain.Flag{
				Name:      valFoo,
				Usage:     valUsage,
				Default:   "default value",
				InputEnvs: []string{envFOO},
			},
			exp: &cli.StringFlag{
				Name:    valFoo,
				Usage:   valUsage,
				Value:   "default value",
				EnvVars: []string{envFOO},
			},
		},
		{
			title: "required",
			flag: domain.Flag{
				Name:      valFoo,
				Usage:     valUsage,
				InputEnvs: []string{envFOO},
				Required:  true,
			},
			exp: &cli.StringFlag{
				Name:    valFoo,
				Usage:   valUsage,
				EnvVars: []string{envFOO},
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
		task  domain.Task
		exp   cli.Command
	}{
		{
			title: "no flag",
			task: domain.Task{
				Name:        valTest,
				Short:       "t",
				Usage:       valUsage,
				Description: valDescription,
			},
			exp: cli.Command{
				Name:               valTest,
				Aliases:            []string{"t"},
				Usage:              valUsage,
				Description:        valDescription,
				CustomHelpTemplate: cli.CommandHelpTemplate,
				Flags:              []cli.Flag{},
			},
		},
		{
			title: "flag",
			task: domain.Task{
				Name:        valTest,
				Short:       "t",
				Usage:       valUsage,
				Description: valDescription,
				Flags: []domain.Flag{
					{
						Name: valFoo,
					},
				},
			},
			exp: cli.Command{
				Name:               valTest,
				Aliases:            []string{"t"},
				Usage:              valUsage,
				Description:        valDescription,
				CustomHelpTemplate: cli.CommandHelpTemplate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: valFoo,
					},
				},
			},
		},
		{
			title: "args",
			task: domain.Task{
				Name:        valTest,
				Short:       "t",
				Usage:       valUsage,
				Description: valDescription,
				Args: []domain.Arg{
					{
						Name:  valFoo,
						Usage: valUsage,
					},
				},
			},
			exp: cli.Command{
				Name:        valTest,
				Aliases:     []string{"t"},
				Usage:       valUsage,
				Description: valDescription,
				Flags:       []cli.Flag{},
				CustomHelpTemplate: cli.CommandHelpTemplate + `
ARGUMENTS:
   foo  usage
`,
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			cmd := convertTaskToCommand(d.task, &domain.GlobalFlags{})
			assert.Equal(t, d.exp.Name, cmd.Name)
			assert.Equal(t, d.exp.Aliases, cmd.Aliases)
			assert.Equal(t, d.exp.Usage, cmd.Usage)
			assert.Equal(t, d.exp.Flags, cmd.Flags)
			assert.Equal(t, d.exp.Description, cmd.Description)
			assert.Equal(t, d.exp.CustomHelpTemplate, cmd.CustomHelpTemplate)
		})
	}
}

func Test_setupEnvs(t *testing.T) {
	envs, err := setupEnvs([]string{"{{.name}}-zoo"}, valFoo)
	require.NoError(t, err)
	assert.Equal(t, []string{"FOO_ZOO"}, envs)
}

func Test_setupTask(t *testing.T) {
	data := []struct {
		title string
		task  *domain.Task
		base  *domain.Task
		isErr bool
		exp   *domain.Task
	}{
		{
			title: "flags and args are empty",
			task:  &domain.Task{},
			base:  &domain.Task{},
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
			base: &domain.Task{
				Environment: map[string]string{
					envFOO: valFoo,
				},
			},
			exp: &domain.Task{
				Timeout: domain.Timeout{
					Duration: defaultTimeout,
				},
				Environment: map[string]string{
					envFOO: valFoo,
				},
			},
		},
		{
			title: titleNormal,
			task: &domain.Task{
				Flags: []domain.Flag{
					{
						Name: valFoo,
					},
				},
				Args: []domain.Arg{
					{
						Name: valBar,
					},
				},
				Environment: map[string]string{
					"ZOO":  "zoo",
					envBAR: valBar,
				},
			},
			base: &domain.Task{
				InputEnvs: []string{"{{.name}}"},
				Environment: map[string]string{
					envFOO: valFoo,
					envBAR: valHello,
				},
			},
			exp: &domain.Task{
				Timeout: domain.Timeout{
					Duration: defaultTimeout,
				},
				Flags: []domain.Flag{
					{
						Name:       valFoo,
						InputEnvs:  []string{envFOO},
						ScriptEnvs: []string{},
					},
				},
				Args: []domain.Arg{
					{
						Name:       valBar,
						InputEnvs:  []string{envBAR},
						ScriptEnvs: []string{},
					},
				},
				Environment: map[string]string{
					envFOO: valFoo,
					envBAR: valBar,
					"ZOO":  "zoo",
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := setupTask(d.task, d.base)
			if err != nil {
				if d.isErr {
					return
				}
				assert.Error(t, err)
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
			title: titleNormal,
			cfg: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   valFoo,
						Script: valEnv,
					},
				},
			},
			exp: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   valFoo,
						Script: valEnv,
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
		t.Run(d.title, func(t *testing.T) {
			err := setupConfig(d.cfg)
			if err != nil {
				if d.isErr {
					return
				}
				assert.Error(t, err)
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
			title: titleNormal,
			cfg: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   valFoo,
						Script: valEnv,
					},
				},
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(_ *testing.T) {
			app := cli.NewApp()
			updateAppWithConfig(app, d.cfg, &domain.GlobalFlags{WorkingDir: "/tmp"})
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
			txt:   valHello,
			exp:   valHello,
		},
	}
	for _, d := range data {
		assert.Equal(t, d.exp, getHelp(d.txt, d.task))
	}
}
