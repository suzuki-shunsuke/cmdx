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
				Name:  "foo",
				Short: "f",
				Usage: "usage",
				Env:   "FOO",
				Type:  "bool",
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
				Name:    "foo",
				Usage:   "usage",
				Default: "default value",
				Env:     "FOO",
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
				Env:      "FOO",
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
			cmd := convertTaskToCommand(d.task)
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
