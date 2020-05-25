package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

func Test_value(t *testing.T) {
	data := []struct {
		title    string
		val      string
		validate domain.Validate
		isErr    bool
	}{
		{
			title:    "no validation",
			val:      "foo",
			validate: domain.Validate{},
			isErr:    false,
		},
		{
			title: "int",
			val:   "0",
			validate: domain.Validate{
				Type: "int",
			},
			isErr: false,
		},
		{
			title: "int error",
			val:   "foo",
			validate: domain.Validate{
				Type: "int",
			},
			isErr: true,
		},
		{
			title: "url",
			val:   "http://example.com",
			validate: domain.Validate{
				Type: "url",
			},
			isErr: false,
		},
		{
			title: "url error",
			val:   "foo",
			validate: domain.Validate{
				Type: "url",
			},
			isErr: true,
		},
		{
			title: "email",
			val:   "foo@example.com",
			validate: domain.Validate{
				Type: "email",
			},
			isErr: false,
		},
		{
			title: "email error",
			val:   "foo",
			validate: domain.Validate{
				Type: "email",
			},
			isErr: true,
		},
		{
			title: "contain",
			val:   "hello",
			validate: domain.Validate{
				Contain: "lo",
			},
			isErr: false,
		},
		{
			title: "contain error",
			val:   "hello",
			validate: domain.Validate{
				Contain: "foo",
			},
			isErr: true,
		},
		{
			title: "prefix",
			val:   "hello",
			validate: domain.Validate{
				Prefix: "hel",
			},
			isErr: false,
		},
		{
			title: "prefix error",
			val:   "hello",
			validate: domain.Validate{
				Prefix: "ell",
			},
			isErr: true,
		},
		{
			title: "suffix",
			val:   "hello",
			validate: domain.Validate{
				Suffix: "lo",
			},
			isErr: false,
		},
		{
			title: "suffix error",
			val:   "hello",
			validate: domain.Validate{
				Suffix: "ll",
			},
			isErr: true,
		},
		{
			title: "min length",
			val:   "hello",
			validate: domain.Validate{
				MinLength: 3,
			},
			isErr: false,
		},
		{
			title: "min length error",
			val:   "hello",
			validate: domain.Validate{
				MinLength: 6,
			},
			isErr: true,
		},
		{
			title: "max length",
			val:   "hello",
			validate: domain.Validate{
				MaxLength: 5,
			},
			isErr: false,
		},
		{
			title: "max length error",
			val:   "hello",
			validate: domain.Validate{
				MaxLength: 4,
			},
			isErr: true,
		},
		{
			title: "enum",
			val:   "hello",
			validate: domain.Validate{
				Enum: []string{"bar", "hello"},
			},
			isErr: false,
		},
		{
			title: "enum error",
			val:   "hello",
			validate: domain.Validate{
				Enum: []string{"bar", "zoo"},
			},
			isErr: true,
		},
		{
			title: "regexp",
			val:   "hello",
			validate: domain.Validate{
				RegExp: "^h.llo",
			},
			isErr: false,
		},
		{
			title: "regexp error",
			val:   "hello",
			validate: domain.Validate{
				RegExp: "^ho",
			},
			isErr: true,
		},
		{
			title: "invalid regexp",
			val:   "hello",
			validate: domain.Validate{
				RegExp: "(.",
			},
			isErr: true,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			err := value(d.val, d.validate)
			if d.isErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}
