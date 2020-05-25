package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestStrList_UnmarshalYAML(t *testing.T) {
	data := []struct {
		title string
		isErr bool
		src   string
		exp   StrList
	}{
		{
			title: "string",
			src:   `"foo"`,
			exp:   StrList{"foo"},
		},
		{
			title: "list",
			src:   `["foo", "bar"]`,
			exp:   StrList{"foo", "bar"},
		},
		{
			title: "error",
			src:   `"tru`,
			isErr: true,
		},
	}
	for _, d := range data {
		list := StrList{}
		err := yaml.Unmarshal([]byte(d.src), &list)
		if d.isErr {
			assert.NotNil(t, err)
			return
		}
		assert.Nil(t, err)
		assert.Equal(t, d.exp, list)
	}
}
