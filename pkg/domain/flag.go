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
	Name       string        `json:"name"`
	Short      string        `json:"short,omitempty"`
	Usage      string        `json:"usage,omitempty"`
	Default    string        `json:"default,omitempty"`
	InputEnvs  []string      `json:"input_envs,omitempty" yaml:"input_envs"`
	ScriptEnvs []string      `json:"script_envs,omitempty" yaml:"script_envs"`
	Type       string        `json:"type,omitempty"`
	Required   bool          `json:"required,omitempty"`
	Prompt     prompt.Prompt `json:"prompt,omitempty"`
	Validate   []Validate    `json:"validate,omitempty"`
}

type Task struct {
	Name        string            `json:"name"`
	Short       string            `json:"short,omitempty"`
	Description string            `json:"description,omitempty"`
	Usage       string            `json:"usage,omitempty"`
	Flags       []Flag            `json:"flags,omitempty"`
	Args        []Arg             `json:"args,omitempty"`
	InputEnvs   []string          `json:"input_envs,omitempty" yaml:"input_envs"`
	ScriptEnvs  []string          `json:"script_envs,omitempty" yaml:"script_envs"`
	Environment map[string]string `json:"environment,omitempty"`
	Script      string            `json:"script"`
	Timeout     Timeout           `json:"timeout,omitempty"`
	Require     Require           `json:"require,omitempty"`
	Quiet       *bool             `json:"quiet,omitempty"`
	Shell       []string          `json:"shell,omitempty"`
	Tasks       []Task            `json:"tasks,omitempty"`
}

type Arg struct {
	Name       string        `json:"name"`
	Usage      string        `json:"usage,omitempty"`
	Default    string        `json:"default,omitempty"`
	InputEnvs  []string      `json:"input_envs,omitempty" yaml:"input_envs"`
	ScriptEnvs []string      `json:"script_envs,omitempty" yaml:"script_envs"`
	Required   bool          `json:"required,omitempty"`
	Prompt     prompt.Prompt `json:"prompt,omitempty"`
	Validate   []Validate    `json:"validate,omitempty"`
}

type Require struct {
	Exec        []StrList `json:"exec,omitempty"`
	Environment []StrList `json:"environment,omitempty"`
}

type Timeout struct {
	Duration  int `json:"duration,omitempty"`
	KillAfter int `json:"kill_after,omitempty" yaml:"kill_after"`
}

type Config struct {
	Tasks       []Task            `json:"tasks"`
	InputEnvs   []string          `json:"input_envs,omitempty" yaml:"input_envs"`
	ScriptEnvs  []string          `json:"script_envs,omitempty" yaml:"script_envs"`
	Environment map[string]string `json:"environment,omitempty"`
	Timeout     Timeout           `json:"timeout,omitempty"`
	Quiet       *bool             `json:"quiet,omitempty"`
}

type Validate struct {
	Type      string   `json:"type,omitempty"`
	RegExp    string   `json:"regexp,omitempty" yaml:"regexp"`
	MinLength int      `json:"min_length,omitempty" yaml:"min_length"`
	MaxLength int      `json:"max_length,omitempty" yaml:"max_length"`
	Prefix    string   `json:"prefix,omitempty"`
	Suffix    string   `json:"suffix,omitempty"`
	Contain   string   `json:"contain,omitempty"`
	Enum      []string `json:"enum,omitempty"`

	Min int `json:"min,omitempty"`
	Max int `json:"max,omitempty"`
}

type HasIsSet interface {
	IsSet(k string) bool
}
