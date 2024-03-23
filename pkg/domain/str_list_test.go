package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
			assert.Error(t, err)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, d.exp, list)
	}
}
