package domain

import (
	"github.com/suzuki-shunsuke/cmdx/pkg/prompt"
)

type GlobalFlags struct {
	DryRun     bool
	Quiet      *bool
	WorkingDir string
}

type Flag struct {
	Name       string
	Short      string
	Usage      string
	Default    string
	InputEnvs  []string `yaml:"input_envs"`
	ScriptEnvs []string `yaml:"script_envs"`
	Type       string
	Required   bool
	Prompt     prompt.Prompt
	Validate   []Validate
}

type Task struct {
	Name        string
	Short       string
	Description string
	Usage       string
	Flags       []Flag
	Args        []Arg
	InputEnvs   []string `yaml:"input_envs"`
	ScriptEnvs  []string `yaml:"script_envs"`
	Environment map[string]string
	Script      string
	Timeout     Timeout
	Require     Require
	Quiet       *bool
	Shell       []string
}

type Arg struct {
	Name       string
	Usage      string
	Default    string
	InputEnvs  []string `yaml:"input_envs"`
	ScriptEnvs []string `yaml:"script_envs"`
	Required   bool
	Prompt     prompt.Prompt
	Validate   []Validate
}

type Require struct {
	Exec        []StrList
	Environment []StrList
}

type Timeout struct {
	Duration  int
	KillAfter int `yaml:"kill_after"`
}

type Config struct {
	Tasks       []Task
	InputEnvs   []string `yaml:"input_envs"`
	ScriptEnvs  []string `yaml:"script_envs"`
	Environment map[string]string
	Timeout     Timeout
	Quiet       *bool
}

type Validate struct {
	Type      string
	RegExp    string `yaml:"regexp"`
	MinLength int    `yaml:"min_length"`
	MaxLength int    `yaml:"max_length"`
	Prefix    string
	Suffix    string
	Contain   string
	Enum      []string

	Min int
	Max int
}

type HasIsSet interface {
	IsSet(string) bool
}
