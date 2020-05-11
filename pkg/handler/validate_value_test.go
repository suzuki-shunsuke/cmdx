package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateValue(t *testing.T) {
	data := []struct {
		title    string
		val      string
		validate Validate
		isErr    bool
	}{
		{
			title:    "no validation",
			val:      "foo",
			validate: Validate{},
			isErr:    false,
		},
		{
			title: "int",
			val:   "0",
			validate: Validate{
				Type: "int",
			},
			isErr: false,
		},
		{
			title: "int error",
			val:   "foo",
			validate: Validate{
				Type: "int",
			},
			isErr: true,
		},
		{
			title: "url",
			val:   "http://example.com",
			validate: Validate{
				Type: "url",
			},
			isErr: false,
		},
		{
			title: "url error",
			val:   "foo",
			validate: Validate{
				Type: "url",
			},
			isErr: true,
		},
		{
			title: "email",
			val:   "foo@example.com",
			validate: Validate{
				Type: "email",
			},
			isErr: false,
		},
		{
			title: "email error",
			val:   "foo",
			validate: Validate{
				Type: "email",
			},
			isErr: true,
		},
		{
			title: "contain",
			val:   "hello",
			validate: Validate{
				Contain: "lo",
			},
			isErr: false,
		},
		{
			title: "contain error",
			val:   "hello",
			validate: Validate{
				Contain: "foo",
			},
			isErr: true,
		},
		{
			title: "prefix",
			val:   "hello",
			validate: Validate{
				Prefix: "hel",
			},
			isErr: false,
		},
		{
			title: "prefix error",
			val:   "hello",
			validate: Validate{
				Prefix: "ell",
			},
			isErr: true,
		},
		{
			title: "suffix",
			val:   "hello",
			validate: Validate{
				Suffix: "lo",
			},
			isErr: false,
		},
		{
			title: "suffix error",
			val:   "hello",
			validate: Validate{
				Suffix: "ll",
			},
			isErr: true,
		},
		{
			title: "min length",
			val:   "hello",
			validate: Validate{
				MinLength: 3,
			},
			isErr: false,
		},
		{
			title: "min length error",
			val:   "hello",
			validate: Validate{
				MinLength: 6,
			},
			isErr: true,
		},
		{
			title: "max length",
			val:   "hello",
			validate: Validate{
				MaxLength: 5,
			},
			isErr: false,
		},
		{
			title: "max length error",
			val:   "hello",
			validate: Validate{
				MaxLength: 4,
			},
			isErr: true,
		},
		{
			title: "enum",
			val:   "hello",
			validate: Validate{
				Enum: []string{"bar", "hello"},
			},
			isErr: false,
		},
		{
			title: "enum error",
			val:   "hello",
			validate: Validate{
				Enum: []string{"bar", "zoo"},
			},
			isErr: true,
		},
		{
			title: "regexp",
			val:   "hello",
			validate: Validate{
				RegExp: "^h.llo",
			},
			isErr: false,
		},
		{
			title: "regexp error",
			val:   "hello",
			validate: Validate{
				RegExp: "^ho",
			},
			isErr: true,
		},
		{
			title: "invalid regexp",
			val:   "hello",
			validate: Validate{
				RegExp: "(.",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := validateValue(d.val, d.validate)
			if d.isErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}
