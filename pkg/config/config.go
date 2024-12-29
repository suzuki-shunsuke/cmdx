package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/suzuki-shunsuke/go-cliutil"
	"gopkg.in/yaml.v3"
)

const configurationFileTemplate = `---
# the configuration file of cmdx, which is a task runner.
# https://github.com/suzuki-shunsuke/cmdx
# timeout:
#   duration: 600
#   kill_after: 30
# input_envs:
# - "{{.name}}"
# script_envs:
# - "{{.name}}"
# environment:
#   FOO: foo
tasks:
- name: hello
  # short: h
  description: hello task
  flags:
  # - name: source
  #   short: s
  #   usage: source file path
  #   description: source file path
  #   default: .drone.jsonnet
  #   required: true
  # - name: force
  #   short: f
  #   usage: force
  #   type: bool
  args:
  # - name: name
  #   usage: source file path
  #   required: true
  #   input_envs:
  #   - NAME
  environment:
    FOO: foo
  script: "echo $FOO"
`

type Client struct{}

func New() *Client {
	return &Client{}
}

func (client *Client) Read(cfgFilePath string, cfg any) error {
	f, err := os.Open(cfgFilePath)
	if err != nil {
		return fmt.Errorf("failed to open the configuration file %s: %w", cfgFilePath, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read the configuration file %s: %w", cfgFilePath, err)
	}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("failed to parse the configuration file. the configuration file is invalid: %s: %w", cfgFilePath, err)
	}
	return nil
}

func (client *Client) Create(p string) error {
	if _, err := os.Stat(p); err == nil {
		// If the configuration file already exists, do nothing.
		return nil
	}
	if err := os.WriteFile(p, []byte(configurationFileTemplate), 0o644); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("failed to create the configuration file %s: %w", p, err)
	}
	return nil
}

func (client *Client) GetFilePath(cfgFileName string) (string, error) {
	names := []string{".cmdx.yaml", ".cmdx.yml", "cmdx.yaml", "cmdx.yml"}
	if cfgFileName != "" {
		names = []string{cfgFileName}
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get the current directory path: %w", err)
	}
	p, err := cliutil.FindFile(wd, func(name string) bool {
		_, err := os.Stat(name)
		return err == nil
	}, names...)
	if err == nil {
		return p, nil
	}
	return "", errors.New("the configuration file is not found")
}
