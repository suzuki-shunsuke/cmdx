package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateUniqueName(t *testing.T) {
	data := []struct {
		title    string
		name     string
		names    map[string]struct{}
		exp      bool
		expNames map[string]struct{}
	}{
		{
			title: "normal",
			name:  "foo",
			names: map[string]struct{}{},
			exp:   true,
			expNames: map[string]struct{}{
				"foo": {},
			},
		},
		{
			title: "normal 2",
			name:  "foo",
			names: map[string]struct{}{
				"bar": {},
			},
			exp: true,
			expNames: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
		},
		{
			title: "false",
			name:  "foo",
			names: map[string]struct{}{
				"foo": {},
			},
			exp: false,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			f := validateUniqueName(d.name, d.names)
			if d.exp {
				assert.True(t, f)
				assert.Equal(t, d.expNames, d.names)
				return
			}
			assert.False(t, f)
		})
	}
}

func Test_validateConfig(t *testing.T) {
	data := []struct {
		title string
		cfg   *Config
		isErr bool
	}{
		{
			title: "no args and flags",
			cfg:   &Config{},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := validateConfig(d.cfg)
			if err == nil {
				assert.False(t, d.isErr)
				return
			}
			if d.isErr {
				return
			}
			assert.NotNil(t, err)
		})
	}
}
