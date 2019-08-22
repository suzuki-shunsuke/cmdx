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
				Name:     "foo",
				Usage:    "usage",
				EnvVar:   "FOO",
				Required: true,
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			assert.Equal(t, d.exp, newFlag(d.flag))
		})
	}
}
