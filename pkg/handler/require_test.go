package handler

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_requireExec(t *testing.T) {
	data := []struct {
		title string
		execs []StrList
		isErr bool
	}{
		{
			title: "no validation",
			execs: nil,
			isErr: false,
		},
		{
			title: "no validation 2",
			execs: []StrList{nil},
			isErr: false,
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			err := requireExec(d.execs)
			if d.isErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func Test_requireEnv(t *testing.T) {
	data := []struct {
		title   string
		envs    []StrList
		setEnvs map[string]string
		isErr   bool
	}{
		{
			title: "no validation",
			envs:  nil,
			isErr: false,
		},
		{
			title: "ok",
			envs: []StrList{
				{},
				{"FOO"},
				{"BAR", "ZOO"},
			},
			setEnvs: map[string]string{
				"FOO": "foo",
				"BAR": "bar",
				"ZOO": "zoo",
			},
			isErr: false,
		},
		{
			title: "FOO is required",
			envs: []StrList{
				{"FOO"},
			},
			setEnvs: nil,
			isErr:   true,
		},
		{
			title: "FOO or BAR is required",
			envs: []StrList{
				{"FOO", "BAR"},
			},
			setEnvs: nil,
			isErr:   true,
		},
	}
	defer os.Clearenv()
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			os.Clearenv()
			for k, v := range d.setEnvs {
				os.Setenv(k, v)
			}
			err := requireEnv(d.envs)
			if d.isErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
		})
	}
}
