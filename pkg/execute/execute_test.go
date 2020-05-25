package execute

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutor_Run(t *testing.T) {
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
	ctx := context.Background()
	exc := New()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			err := exc.Run(ctx, d.params)
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
