package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/suzuki-shunsuke/go-cliutil"
	"github.com/urfave/cli/v2"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	"github.com/suzuki-shunsuke/cmdx/pkg/signal"
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

	rootHelp = `cmdx - task runner
https://github.com/suzuki-shunsuke/cmdx

Configuration file isn't found.
First of all, let's create a configuration file.

$ cmdx --init

Or if the configuration file already exists but the file path is unusual, please specify the path by --config (-c) option.

$ cmdx -c <YOUR_CONFIGURATION_FILE_PATH> <COMMAND> ...
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
		Quiet       *bool
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
		Quiet       *bool
	}

	Require struct {
		Exec        []StrList
		Environment []StrList
	}

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
		Validate   []Validate
	}

	Arg struct {
		Name       string
		Usage      string
		Default    string
		InputEnvs  []string `yaml:"input_envs"`
		ScriptEnvs []string `yaml:"script_envs"`
		Required   bool
		Prompt     Prompt
		Validate   []Validate
	}

	Validate struct {
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
)

func Main(args []string) error {
	app := cli.NewApp()
	setupApp(app)
	app.HideHelp = true
	app.BashComplete = rootBashCompletion(args)

	app.Action = mainAction(args)

	app.CustomAppHelpTemplate = rootHelp
	c, cancel := context.WithCancel(context.Background())
	defer cancel()
	go signal.Handle(cancel)
	return app.RunContext(c, args)
}

func mainAction(args []string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		cfg := Config{}
		cfgFilePath := c.String("config")
		initFlag := c.Bool("init")
		listFlag := c.Bool("list")
		helpFlag := c.Bool("help")
		workingDirFlag := c.String("working-dir")
		cfgFileName := c.String("name")
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
		var quiet *bool
		if c.IsSet("quiet") {
			q := c.Bool("quiet")
			quiet = &q
		}
		updateAppWithConfig(app, &cfg, &GlobalFlags{
			DryRun:     c.Bool("dry-run"),
			Quiet:      quiet,
			WorkingDir: workingDirFlag,
		})
		return app.Run(args)
	}
}

type GlobalFlags struct {
	DryRun     bool
	Quiet      *bool
	WorkingDir string
}

func setupApp(app *cli.App) {
	app.Name = "cmdx"
	app.Version = domain.Version
	app.Authors = []*cli.Author{
		{
			Name: "Shunsuke Suzuki",
		},
	}
	app.Usage = appUsage
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "configuration file path",
			EnvVars: []string{"CMDX_CONFIG_PATH"},
		},
		&cli.StringFlag{
			Name:    "name",
			Aliases: []string{"n"},
			Usage:   "configuration file name. The configuration file is searched from the current directory to the root directory recursively",
		},
		&cli.StringFlag{
			Name:    "working-dir",
			Aliases: []string{"w"},
			Usage:   "The working directory path. By default, the task is run on the directory where the configuration file is found",
			EnvVars: []string{"CMDX_WORKING_DIR"},
		},
		&cli.BoolFlag{
			Name:    "init",
			Aliases: []string{"i"},
			Usage:   "create the configuration file",
		},
		&cli.BoolFlag{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list tasks",
		},
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "don't output the executed command",
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"d"},
			Usage:   "output the script but don't run it actually",
		},
	}
}

func newFlag(flag Flag) cli.Flag {
	name := flag.Name
	if flag.Short != "" {
		name += ", " + flag.Short
	}
	switch flag.Type {
	case boolFlagType:
		return &cli.BoolFlag{
			Name:    name,
			Usage:   flag.Usage,
			EnvVars: flag.InputEnvs,
		}
	default:
		return &cli.StringFlag{
			Name:    name,
			Usage:   flag.Usage,
			Value:   flag.Default,
			EnvVars: flag.InputEnvs,
		}
	}
}

func getHelp(txt string, task Task) string {
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
		txt = strings.Replace(
			txt, "[arguments...]", strings.Join(argNames, " "), 1) + `
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
			txt += `
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
			txt += `

REQUIRED ENVIRONMENT VARIABLES:
` + strings.Join(helps, "\n")
		}
	}
	return txt
}

func convertTaskToCommand(task Task, gFlags *GlobalFlags) *cli.Command {
	flags := make([]cli.Flag, len(task.Flags))
	for j, flag := range task.Flags {
		flags[j] = newFlag(flag)
	}
	help := getHelp(cli.CommandHelpTemplate, task)

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

	return &cli.Command{
		Name:               task.Name,
		Aliases:            []string{task.Short},
		Usage:              task.Usage,
		Description:        task.Description,
		Flags:              flags,
		Action:             newCommandAction(task, gFlags, scriptEnvs),
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
			if err := validateValueWithValidates(val, arg.Validate); err != nil {
				return fmt.Errorf(arg.Name+" is invalid: %w", err)
			}
			continue
		}
		// the positional argument isn't given
		isBoundEnv := false
		for _, e := range arg.InputEnvs {
			if v, ok := os.LookupEnv(e); ok {
				isBoundEnv = true
				vars[arg.Name] = v
				if err := validateValueWithValidates(v, arg.Validate); err != nil {
					return fmt.Errorf(arg.Name+" is invalid: %w", err)
				}
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
			if v, ok := val.(string); ok {
				if err := validateValueWithValidates(v, arg.Validate); err != nil {
					return fmt.Errorf(arg.Name+" is invalid: %w", err)
				}
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
	task Task, gFlags *GlobalFlags, scriptEnvs map[string][]string,
) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// create vars and envs
		// run command
		if err := requireExec(task.Require.Exec); err != nil {
			return err
		}
		if err := requireEnv(task.Require.Environment); err != nil {
			return err
		}

		vars := map[string]interface{}{}

		// get flag values and set them to vars
		if err := setFlagValues(c, task.Flags, vars); err != nil {
			return err
		}

		// get args and set them to vars
		if err := updateVarsByArgs(task.Args, c.Args().Slice(), vars); err != nil {
			return err
		}

		// update environment variables which are set to script
		envs := bindScriptEnvs(os.Environ(), vars, scriptEnvs)

		for k, v := range task.Environment {
			envs = append(envs, k+"="+v)
		}

		scr, err := renderTemplate(task.Script, vars)
		if err != nil {
			return fmt.Errorf("failed to parse the script - %s: %w", task.Script, err)
		}

		quiet := false
		if gFlags.Quiet != nil {
			quiet = *gFlags.Quiet
		} else if task.Quiet != nil {
			quiet = *task.Quiet
		}

		return runScript(
			c.Context, scr, gFlags.WorkingDir, envs, task.Timeout, quiet, gFlags.DryRun)
	}
}

func updateAppWithConfig(app *cli.App, cfg *Config, gFlags *GlobalFlags) {
	cmds := make([]*cli.Command, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		cmds[i] = convertTaskToCommand(task, gFlags)
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

	if task.Quiet == nil {
		task.Quiet = cfg.Quiet
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
		task := task
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
