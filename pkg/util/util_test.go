package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate(t *testing.T) {
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
			s, err := RenderTemplate(d.base, d.data)
			if err != nil {
				if d.isErr {
					return
				}
				assert.Error(t, err)
				return
			}
			assert.Equal(t, d.exp, s)
		})
	}
}
