package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_runScript(t *testing.T) {
	data := []struct {
		title  string
		script string
		wd     string
		envs   []string
		quiet  bool
		dryRun bool
		isErr  bool
	}{
		{
			title:  "dry run",
			script: "true",
			dryRun: true,
		},
		{
			title:  "normal",
			script: "true",
		},
		{
			title:  "command is failure",
			script: "false",
			isErr:  true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := runScript(d.script, d.wd, d.envs, d.quiet, d.dryRun)
			if err != nil {
				if d.isErr {
					return
				}
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}
