package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	hasIsSetMock struct {
		flags map[string]struct{}
	}
)

func (h hasIsSetMock) IsSet(flag string) bool {
	_, ok := h.flags[flag]
	return ok
}

func newHasIsSet(flags ...string) hasIsSet {
	m := make(map[string]struct{}, len(flags))
	for _, f := range flags {
		m[f] = struct{}{}
	}
	return hasIsSetMock{
		flags: m,
	}
}

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
			title: "no task",
			cfg:   &Config{},
		},
		{
			title: "normal",
			cfg: &Config{
				Tasks: []Task{
					{
						Name:   "foo",
						Short:  "f",
						Script: "pwd",
					},
				},
			},
		},
		{
			title: "task name duplicates",
			cfg: &Config{
				Tasks: []Task{
					{
						Name:   "foo",
						Script: "pwd",
					},
					{
						Name:   "foo",
						Script: "pwd",
					},
				},
			},
			isErr: true,
		},
		{
			title: "task short name duplicates",
			cfg: &Config{
				Tasks: []Task{
					{
						Name:   "foo",
						Short:  "f",
						Script: "pwd",
					},
					{
						Name:   "bar",
						Short:  "f",
						Script: "pwd",
					},
				},
			},
			isErr: true,
		},
		{
			title: "invalid task",
			cfg: &Config{
				Tasks: []Task{
					{
						Name:  "foo",
						Short: "f",
						Args: []Arg{
							{},
						},
					},
				},
			},
			isErr: true,
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

func Test_validateFlag(t *testing.T) {
	data := []struct {
		title      string
		flag       Flag
		names      map[string]struct{}
		shortNames map[string]struct{}
		isErr      bool
	}{
		{
			title: "name is required",
			isErr: true,
		},
		{
			title: "short name is too long",
			flag: Flag{
				Name:  "foo",
				Short: "foo",
			},
			isErr: true,
		},
		{
			title: "flag name duplicates",
			flag: Flag{
				Name: "foo",
			},
			names: map[string]struct{}{
				"foo": {},
			},
			isErr: true,
		},
		{
			title: "flag short name duplicates",
			flag: Flag{
				Name:  "foo",
				Short: "f",
			},
			shortNames: map[string]struct{}{
				"f": {},
			},
			isErr: true,
		},
		{
			title: "invalid flag type",
			flag: Flag{
				Name: "foo",
				Type: "f",
			},
			isErr: true,
		},
		{
			title: "normal",
			flag: Flag{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			if d.names == nil {
				d.names = map[string]struct{}{}
			}
			if d.shortNames == nil {
				d.shortNames = map[string]struct{}{}
			}
			err := validateFlag("task-name", d.flag, d.names, d.shortNames)
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

func Test_validateArg(t *testing.T) {
	data := []struct {
		title string
		arg   Arg
		names map[string]struct{}
		isErr bool
	}{
		{
			title: "name is required",
			isErr: true,
		},
		{
			title: "arg name duplicates",
			arg: Arg{
				Name: "foo",
			},
			names: map[string]struct{}{
				"foo": {},
			},
			isErr: true,
		},
		{
			title: "normal",
			arg: Arg{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			if d.names == nil {
				d.names = map[string]struct{}{}
			}
			err := validateArg("task-name", d.arg, d.names)
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

func Test_validateTask(t *testing.T) {
	data := []struct {
		title string
		task  Task
		isErr bool
	}{
		{
			title: "name is required",
			isErr: true,
		},
		{
			title: "invalid flag",
			task: Task{
				Name: "foo",
				Flags: []Flag{
					{},
				},
			},
			isErr: true,
		},
		{
			title: "invalid arg",
			task: Task{
				Name: "foo",
				Args: []Arg{
					{},
				},
			},
			isErr: true,
		},
		{
			title: "normal",
			task: Task{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := validateTask(d.task)
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

func Test_validateFlagRequired(t *testing.T) {
	data := []struct {
		title   string
		flagSet []string
		flags   []Flag
		isErr   bool
	}{
		{
			title:   "normal",
			flagSet: []string{"bar"},
			flags: []Flag{
				{
					Name: "foo",
				},
				{
					Name:     "bar",
					Required: true,
				},
			},
		},
		{
			title: "required",
			flags: []Flag{
				{
					Name:     "foo",
					Required: true,
				},
			},
			isErr: true,
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			c := newHasIsSet(d.flagSet...)
			err := validateFlagRequired(c, d.flags)
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
