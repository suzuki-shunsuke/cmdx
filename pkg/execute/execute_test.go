package execute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutor_Run(t *testing.T) {
	t.Parallel()
	data := []struct {
		title  string
		params *Params
		isErr  bool
	}{
		{
			title: "dry run",
			params: &Params{
				Script:  "true",
				DryRun:  true,
				Timeout: &Timeout{},
			},
		},
		{
			title: "normal",
			params: &Params{
				Script:  "true",
				Timeout: &Timeout{},
			},
		},
		{
			title: "command is failure",
			isErr: true,
			params: &Params{
				Script:  "false",
				Timeout: &Timeout{},
			},
		},
	}
	exc := New()
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			t.Parallel()
			err := exc.Run(t.Context(), d.params)
			if err != nil {
				if d.isErr {
					return
				}
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
