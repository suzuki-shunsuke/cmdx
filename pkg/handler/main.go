package handler

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/suzuki-shunsuke/go-cliutil"
	"github.com/urfave/cli"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

const (
	boolFlagType      = "bool"
	confirmPromptType = "confirm"
	defaultTimeout    = 36000 // default 10H

	configurationFileTemplate = `---
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

	appUsage = "task runner"
)

type (
	Config struct {
		Tasks       []Task
		InputEnvs   []string `yaml:"input_envs"`
		ScriptEnvs  []string `yaml:"script_envs"`
		Environment map[string]string
		Timeout     Timeout
	}

	Task struct {
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
	}

	Require struct {
		Exec        []StrList
		Environment []StrList
	}

	StrList []string

	Timeout struct {
		Duration  int
		KillAfter int `yaml:"kill_after"`
	}

	Prompt struct {
		Type    string
		Message string
		Help    string
		Options []string
	}

	Flag struct {
		Name       string
		Short      string
		Usage      string
		Default    string
		InputEnvs  []string `yaml:"input_envs"`
		ScriptEnvs []string `yaml:"script_envs"`
		Type       string
		Required   bool
		Prompt     Prompt
	}

	Arg struct {
		Name       string
		Usage      string
		Default    string
		InputEnvs  []string `yaml:"input_envs"`
		ScriptEnvs []string `yaml:"script_envs"`
		Required   bool
		Prompt     Prompt
	}
)

func (list *StrList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val interface{}
	if err := unmarshal(&val); err != nil {
		return err
	}
	if s, ok := val.(string); ok {
		*list = []string{s}
		return nil
	}
	if intfArr, ok := val.([]interface{}); ok {
		strArr := make([]string, len(intfArr))
		for i, intf := range intfArr {
			if s, ok := intf.(string); ok {
				strArr[i] = s
				continue
			}
			return fmt.Errorf("the type of the value must be string: %v", intf)
		}
		*list = strArr
	}
	return nil
}

func Main() error {
	app := cli.NewApp()
	app.HideHelp = true
	setupApp(app)

	app.Action = mainAction

	app.CustomAppHelpTemplate = `cmdx - task runner
https://github.com/suzuki-shunsuke/cmdx

Configuration file isn't found.
First of all, let's create a configuration file.

$ cmdx --init

Or if the configuration file already exists but the file path is unusual, please specify the path by --config (-c) option.

$ cmdx -c <YOUR_CONFIGURATION_FILE_PATH> <COMMAND> ...
`
	return app.Run(os.Args)
}

func mainAction(c *cli.Context) error {
	cfg := Config{}
	cfgFilePath := c.GlobalString("config")
	initFlag := c.GlobalBool("init")
	listFlag := c.GlobalBool("list")
	helpFlag := c.GlobalBool("help")
	workingDirFlag := c.GlobalString("working-dir")
	cfgFileName := c.GlobalString("name")
	if initFlag {
		if cfgFilePath != "" {
			return createConfigFile(cfgFilePath)
		}
		if cfgFileName != "" {
			return createConfigFile(cfgFileName)
		}
		return createConfigFile(".cmdx.yaml")
	}

	if cfgFilePath == "" {
		var err error
		cfgFilePath, err = getConfigFilePath(cfgFileName)
		if err != nil {
			if helpFlag && cfgFileName == "" {
				return cli.ShowAppHelp(c)
			}
			if c.NArg() == 1 && c.Args().First() == "help" && cfgFileName == "" {
				return cli.ShowAppHelp(c)
			}
			if c.NArg() == 1 && c.Args().First() == "version" && cfgFileName == "" {
				cli.ShowVersion(c)
				return nil
			}
			return err
		}
	}

	if err := readConfig(cfgFilePath, &cfg); err != nil {
		return err
	}
	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("please fix the configuration file: %w", err)
	}

	if err := setupConfig(&cfg); err != nil {
		return err
	}

	if listFlag {
		arr := make([]string, len(cfg.Tasks))
		for i, task := range cfg.Tasks {
			name := task.Name
			if task.Short != "" {
				name += ", " + task.Short
			}
			arr[i] = name + " - " + task.Usage
		}
		fmt.Println(strings.Join(arr, "\n"))
		return nil
	}

	app := cli.NewApp()
	setupApp(app)
	if workingDirFlag == "" {
		workingDirFlag = filepath.Dir(cfgFilePath)
	}
	updateAppWithConfig(app, &cfg, workingDirFlag)
	return app.Run(os.Args)
}

func setupApp(app *cli.App) {
	app.Name = "cmdx"
	app.Version = domain.Version
	app.Authors = []cli.Author{
		{
			Name: "Shunsuke Suzuki",
		},
	}
	app.Usage = appUsage

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, c",
			Usage:  "configuration file path",
			EnvVar: "CMDX_CONFIG_PATH",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "configuration file name. The configuration file is searched from the current directory to the root directory recursively",
		},
		cli.StringFlag{
			Name:   "working-dir, w",
			Usage:  "The working directory path. By default, the task is run on the directory where the configuration file is found",
			EnvVar: "CMDX_WORKING_DIR",
		},
		cli.BoolFlag{
			Name:  "init, i",
			Usage: "create the configuration file",
		},
		cli.BoolFlag{
			Name:  "list, l",
			Usage: "list tasks",
		},
		cli.BoolFlag{
			Name:  "help, h",
			Usage: "show help",
		},
		cli.BoolFlag{
			Name:  "quiet, q",
			Usage: "don't output the executed command",
		},
		cli.BoolFlag{
			Name:  "dry-run, d",
			Usage: "output the script but don't run it actually",
		},
	}
}

func newFlag(flag Flag) cli.Flag {
	name := flag.Name
	if flag.Short != "" {
		name += ", " + flag.Short
	}
	env := strings.Join(flag.InputEnvs, ",")
	switch flag.Type {
	case boolFlagType:
		return cli.BoolFlag{
			Name:   name,
			Usage:  flag.Usage,
			EnvVar: env,
		}
	default:
		return cli.StringFlag{
			Name:   name,
			Usage:  flag.Usage,
			Value:  flag.Default,
			EnvVar: env,
		}
	}
}

func convertTaskToCommand(task Task, wd string) cli.Command {
	flags := make([]cli.Flag, len(task.Flags))
	for j, flag := range task.Flags {
		flags[j] = newFlag(flag)
	}
	help := cli.CommandHelpTemplate
	if len(task.Args) != 0 {
		argHelps := make([]string, len(task.Args))
		argNames := make([]string, len(task.Args))
		for i, arg := range task.Args {
			h := "   " + arg.Name
			if arg.Usage != "" {
				h += "  " + arg.Usage
			}
			argHelps[i] = h
			argNames[i] = "<" + arg.Name + ">"
		}
		help = strings.Replace(
			cli.CommandHelpTemplate, "[arguments...]", strings.Join(argNames, " "), 1) + `
ARGUMENTS:
` + strings.Join(argHelps, "\n")
	}

	if len(task.Require.Exec) != 0 {
		helps := make([]string, 0, len(task.Require.Exec))
		for _, require := range task.Require.Exec {
			if len(require) == 0 {
				continue
			}
			helps = append(helps, "  "+strings.Join(require, " or "))
		}
		if len(helps) != 0 {
			help += `
REQUIREMENTS:
` + strings.Join(helps, "\n")
		}
	}

	if len(task.Require.Environment) != 0 {
		helps := make([]string, 0, len(task.Require.Environment))
		for _, require := range task.Require.Environment {
			if len(require) == 0 {
				continue
			}
			helps = append(helps, "  "+strings.Join(require, " or "))
		}
		if len(helps) != 0 {
			help += `

REQUIRED ENVIRONMENT VARIABLES:
` + strings.Join(helps, "\n")
		}
	}

	scriptEnvs := map[string][]string{}
	for _, flag := range task.Flags {
		if len(flag.ScriptEnvs) != 0 {
			scriptEnvs[flag.Name] = flag.ScriptEnvs
		}
	}
	for _, arg := range task.Args {
		if len(arg.ScriptEnvs) != 0 {
			scriptEnvs[arg.Name] = arg.ScriptEnvs
		}
	}

	return cli.Command{
		Name:               task.Name,
		ShortName:          task.Short,
		Usage:              task.Usage,
		Description:        task.Description,
		Flags:              flags,
		Action:             newCommandAction(task, wd, scriptEnvs),
		CustomHelpTemplate: help,
	}
}

func updateVarsByArgs(
	args []Arg, cArgs []string, vars map[string]interface{},
) error {
	n := len(cArgs)

	for i, arg := range args {
		if i < n {
			val := cArgs[i]
			vars[arg.Name] = val
			continue
		}
		// the positional argument isn't given
		isBoundEnv := false
		for _, e := range arg.InputEnvs {
			if v, ok := os.LookupEnv(e); ok {
				isBoundEnv = true
				vars[arg.Name] = v
				break
			}
		}
		if isBoundEnv {
			continue
		}
		if prompt := createPrompt(arg.Prompt); prompt != nil {
			val, err := getValueByPrompt(prompt, arg.Prompt.Type)
			if err != nil {
				// TODO improvement
				if arg.Default != "" {
					vars[arg.Name] = arg.Default
					continue
				}
				continue
			}
			vars[arg.Name] = val
			continue
		}
		if arg.Default != "" {
			vars[arg.Name] = arg.Default
			continue
		}
		if arg.Required {
			return fmt.Errorf("the %d th argument '%s' is required", i+1, arg.Name)
		}
		vars[arg.Name] = ""
	}

	extraArgs := []string{}
	for i, arg := range cArgs {
		if i < len(args) {
			continue
		}
		extraArgs = append(extraArgs, arg)
	}

	vars["_builtin"] = map[string]interface{}{
		"args":            extraArgs,
		"args_string":     strings.Join(extraArgs, " "),
		"all_args":        cArgs,
		"all_args_string": strings.Join(cArgs, " "),
	}
	return nil
}

func newCommandAction(
	task Task, wd string, scriptEnvs map[string][]string,
) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// create vars and envs
		// run command

		for _, requires := range task.Require.Exec {
			if len(requires) == 0 {
				continue
			}
			f := false
			for _, require := range requires {
				if _, err := exec.LookPath(require); err == nil {
					f = true
					break
				}
			}
			if !f {
				if len(requires) == 1 {
					return errors.New(requires[0] + " is required")
				}
				return errors.New("one of the following is required: " + strings.Join(requires, ", "))
			}
		}

		for _, requires := range task.Require.Environment {
			if len(requires) == 0 {
				continue
			}
			f := false
			for _, require := range requires {
				if os.Getenv(require) != "" {
					f = true
					break
				}
			}
			if !f {
				if len(requires) == 1 {
					return errors.New("the environment variable '" + requires[0] + "' is required")
				}
				return errors.New("one of the following environment variables is required: " + strings.Join(requires, ", "))
			}
		}

		vars := map[string]interface{}{}

		for _, flag := range task.Flags {
			if c.IsSet(flag.Name) {
				var val interface{}
				switch flag.Type {
				case boolFlagType:
					val = c.Bool(flag.Name)
				default:
					val = c.String(flag.Name)
				}
				vars[flag.Name] = val
				continue
			}

			if p := createPrompt(flag.Prompt); p != nil {
				val, err := getValueByPrompt(p, flag.Prompt.Type)
				if err == nil {
					vars[flag.Name] = val
					continue
				}
			}

			switch flag.Type {
			case boolFlagType:
				// don't ues c.Generic if flag.Type == "bool"
				// the value in the template is treated as false
				vars[flag.Name] = c.Bool(flag.Name)
			default:
				if v := c.String(flag.Name); v != "" {
					vars[flag.Name] = v
					continue
				}
				if flag.Required {
					return errors.New(`the flag "` + flag.Name + `" is required`)
				}
				vars[flag.Name] = ""
			}
		}

		err := updateVarsByArgs(task.Args, c.Args(), vars)
		if err != nil {
			return err
		}

		envs := os.Environ()
		for k, envNames := range scriptEnvs {
			switch v := vars[k].(type) {
			case string:
				for _, e := range envNames {
					envs = append(envs, e+"="+v)
				}
			case bool:
				a := strconv.FormatBool(v)
				for _, e := range envNames {
					envs = append(envs, e+"="+a)
				}
			case []string:
				a := strings.Join(v, ",")
				for _, e := range envNames {
					envs = append(envs, e+"="+a)
				}
			}
		}

		for k, v := range task.Environment {
			envs = append(envs, k+"="+v)
		}

		scr, err := renderTemplate(task.Script, vars)
		if err != nil {
			return fmt.Errorf("failed to parse the script - %s: %w", task.Script, err)
		}

		return runScript(
			scr, wd, envs, task.Timeout, c.GlobalBool("quiet"), c.GlobalBool("dry-run"))
	}
}

func updateAppWithConfig(app *cli.App, cfg *Config, wd string) {
	cmds := make([]cli.Command, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		cmds[i] = convertTaskToCommand(task, wd)
	}
	app.Commands = cmds
}

func renderTemplate(base string, data interface{}) (string, error) {
	tmpl, err := template.New("command").Parse(base)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, data)
	return buf.String(), err
}

func setupEnvs(envs []string, name string) ([]string, error) {
	arr := make([]string, len(envs))
	for i, env := range envs {
		e, err := renderTemplate(env, map[string]string{
			"name": name,
		})
		if err != nil {
			return nil, err
		}
		arr[i] = strings.ToUpper(strings.Replace(e, "-", "_", -1))
	}
	return arr, nil
}

func setupTask(task *Task, cfg *Config) error {
	inputEnvs := task.InputEnvs
	if len(inputEnvs) == 0 {
		inputEnvs = cfg.InputEnvs
	}

	scriptEnvs := task.ScriptEnvs
	if len(scriptEnvs) == 0 {
		scriptEnvs = cfg.ScriptEnvs
	}

	if task.Environment == nil {
		task.Environment = map[string]string{}
	}
	for k, v := range cfg.Environment {
		if _, ok := task.Environment[k]; !ok {
			task.Environment[k] = v
		}
	}

	if task.Timeout.Duration == 0 {
		if cfg.Timeout.Duration == 0 {
			task.Timeout.Duration = defaultTimeout
		} else {
			task.Timeout.Duration = cfg.Timeout.Duration
		}
	}

	if task.Timeout.KillAfter == 0 {
		task.Timeout.KillAfter = cfg.Timeout.KillAfter
	}

	for j, flag := range task.Flags {
		if len(flag.InputEnvs) == 0 {
			flag.InputEnvs = inputEnvs
		}
		envs, err := setupEnvs(flag.InputEnvs, flag.Name)
		if err != nil {
			return err
		}
		flag.InputEnvs = envs

		if len(flag.ScriptEnvs) == 0 {
			flag.ScriptEnvs = scriptEnvs
		}
		envs, err = setupEnvs(flag.ScriptEnvs, flag.Name)
		if err != nil {
			return err
		}
		flag.ScriptEnvs = envs
		if flag.Prompt.Type != "" {
			if flag.Prompt.Message == "" {
				flag.Prompt.Message = flag.Name
			}
		}

		task.Flags[j] = flag
	}

	for j, arg := range task.Args {
		if len(arg.InputEnvs) == 0 {
			arg.InputEnvs = inputEnvs
		}
		envs, err := setupEnvs(arg.InputEnvs, arg.Name)
		if err != nil {
			return err
		}
		arg.InputEnvs = envs

		if len(arg.ScriptEnvs) == 0 {
			arg.ScriptEnvs = scriptEnvs
		}
		envs, err = setupEnvs(arg.ScriptEnvs, arg.Name)
		if err != nil {
			return err
		}
		arg.ScriptEnvs = envs

		if arg.Prompt.Type != "" {
			if arg.Prompt.Message == "" {
				arg.Prompt.Message = arg.Name
			}
		}

		task.Args[j] = arg
	}

	return nil
}

func setupConfig(cfg *Config) error {
	for i, task := range cfg.Tasks {
		if err := setupTask(&task, cfg); err != nil {
			return err
		}
		cfg.Tasks[i] = task
	}
	return nil
}

func getConfigFilePath(cfgFileName string) (string, error) {
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
