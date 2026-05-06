package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

const (
	testValFoo            = "foo"
	testValBar            = "bar"
	testValHello          = "hello"
	testValPwd            = "pwd"
	testTitleNormal       = "normal"
	testTitleNameRequired = "name is required"
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
			title: testTitleNormal,
			name:  testValFoo,
			names: map[string]struct{}{},
			exp:   true,
			expNames: map[string]struct{}{
				testValFoo: {},
			},
		},
		{
			title: "normal 2",
			name:  testValFoo,
			names: map[string]struct{}{
				testValBar: {},
			},
			exp: true,
			expNames: map[string]struct{}{
				testValFoo: {},
				testValBar: {},
			},
		},
		{
			title: "false",
			name:  testValFoo,
			names: map[string]struct{}{
				testValFoo: {},
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
			title: testTitleNormal,
			cfg: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   testValFoo,
						Short:  "f",
						Script: testValPwd,
					},
				},
			},
		},
		{
			title: "task name duplicates",
			cfg: &domain.Config{
				Tasks: []domain.Task{
					{
						Name:   testValFoo,
						Script: testValPwd,
					},
					{
						Name:   testValFoo,
						Script: testValPwd,
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
						Name:   testValFoo,
						Short:  "f",
						Script: testValPwd,
					},
					{
						Name:   testValBar,
						Short:  "f",
						Script: testValPwd,
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
						Name:  testValFoo,
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
			title: testTitleNameRequired,
			isErr: true,
		},
		{
			title: "short name is too long",
			flag: domain.Flag{
				Name:  testValFoo,
				Short: testValFoo,
			},
			isErr: true,
		},
		{
			title: "flag name duplicates",
			flag: domain.Flag{
				Name: testValFoo,
			},
			names: map[string]struct{}{
				testValFoo: {},
			},
			isErr: true,
		},
		{
			title: "flag short name duplicates",
			flag: domain.Flag{
				Name:  testValFoo,
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
				Name: testValFoo,
				Type: "f",
			},
			isErr: true,
		},
		{
			title: testTitleNormal,
			flag: domain.Flag{
				Name: testValFoo,
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
			title: testTitleNameRequired,
			isErr: true,
		},
		{
			title: "arg name duplicates",
			arg: domain.Arg{
				Name: testValFoo,
			},
			names: map[string]struct{}{
				testValFoo: {},
			},
			isErr: true,
		},
		{
			title: testTitleNormal,
			arg: domain.Arg{
				Name: testValFoo,
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
			title: testTitleNameRequired,
			isErr: true,
		},
		{
			title: "invalid flag",
			task: domain.Task{
				Name: testValFoo,
				Flags: []domain.Flag{
					{},
				},
			},
			isErr: true,
		},
		{
			title: "invalid arg",
			task: domain.Task{
				Name: testValFoo,
				Args: []domain.Arg{
					{},
				},
			},
			isErr: true,
		},
		{
			title: testTitleNormal,
			task: domain.Task{
				Name: testValFoo,
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
