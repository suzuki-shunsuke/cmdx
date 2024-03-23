package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

func Test_vUniqueName(t *testing.T) {
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
			f := vUniqueName(d.name, d.names)
			if d.exp {
				assert.True(t, f)
				assert.Equal(t, d.expNames, d.names)
				return
			}
			assert.False(t, f)
		})
	}
}

func TestConfig(t *testing.T) {
	data := []struct {
		title string
		cfg   *domain.Config
		isErr bool
	}{
		{
			title: "no task",
			cfg:   &domain.Config{},
		},
		{
			title: "normal",
			cfg: &domain.Config{
				Tasks: []domain.Task{
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
			cfg: &domain.Config{
				Tasks: []domain.Task{
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
			cfg: &domain.Config{
				Tasks: []domain.Task{
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
			cfg: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:  "foo",
						Short: "f",
						Args: []domain.Arg{
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
			err := Config(d.cfg)
			if err == nil {
				assert.False(t, d.isErr)
				return
			}
			if d.isErr {
				return
			}
			assert.Error(t, err)
		})
	}
}

func Test_vFlag(t *testing.T) {
	data := []struct {
		title      string
		flag       domain.Flag
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
			flag: domain.Flag{
				Name:  "foo",
				Short: "foo",
			},
			isErr: true,
		},
		{
			title: "flag name duplicates",
			flag: domain.Flag{
				Name: "foo",
			},
			names: map[string]struct{}{
				"foo": {},
			},
			isErr: true,
		},
		{
			title: "flag short name duplicates",
			flag: domain.Flag{
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
			flag: domain.Flag{
				Name: "foo",
				Type: "f",
			},
			isErr: true,
		},
		{
			title: "normal",
			flag: domain.Flag{
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
			err := vFlag("task-name", d.flag, d.names, d.shortNames)
			if err == nil {
				assert.False(t, d.isErr)
				return
			}
			if d.isErr {
				return
			}
			assert.Error(t, err)
		})
	}
}

func Test_vArg(t *testing.T) {
	data := []struct {
		title string
		arg   domain.Arg
		names map[string]struct{}
		isErr bool
	}{
		{
			title: "name is required",
			isErr: true,
		},
		{
			title: "arg name duplicates",
			arg: domain.Arg{
				Name: "foo",
			},
			names: map[string]struct{}{
				"foo": {},
			},
			isErr: true,
		},
		{
			title: "normal",
			arg: domain.Arg{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			if d.names == nil {
				d.names = map[string]struct{}{}
			}
			err := vArg("task-name", d.arg, d.names)
			if err == nil {
				assert.False(t, d.isErr)
				return
			}
			if d.isErr {
				return
			}
			assert.Error(t, err)
		})
	}
}

func Test_vTask(t *testing.T) {
	data := []struct {
		title string
		task  domain.Task
		isErr bool
	}{
		{
			title: "name is required",
			isErr: true,
		},
		{
			title: "invalid flag",
			task: domain.Task{
				Name: "foo",
				Flags: []domain.Flag{
					{},
				},
			},
			isErr: true,
		},
		{
			title: "invalid arg",
			task: domain.Task{
				Name: "foo",
				Args: []domain.Arg{
					{},
				},
			},
			isErr: true,
		},
		{
			title: "normal",
			task: domain.Task{
				Name: "foo",
			},
		},
	}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := vTask(d.task)
			if err == nil {
				assert.False(t, d.isErr)
				return
			}
			if d.isErr {
				return
			}
			assert.Error(t, err)
		})
	}
}
