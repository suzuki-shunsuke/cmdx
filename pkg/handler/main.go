package handler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
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
		Tasks      []Task
		InputEnvs  []string `yaml:"input_envs"`
		ScriptEnvs []string `yaml:"script_envs"`
		Timeout    Timeout
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
	}

	Timeout struct {
		Duration  time.Duration
		KillAfter time.Duration `yaml:"kill_after"`
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

func Main() error {
	app := cli.NewApp()
	app.HideHelp = true
	setupApp(app)

	app.Action = cliutil.WrapAction(mainAction)

	return app.Run(os.Args)
}

func mainAction(c *cli.Context) error {
	cfg := Config{}
	cfgFilePath := c.GlobalString("config")
	initFlag := c.GlobalBool("init")
	listFlag := c.GlobalBool("list")
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
			return err
		}
	}

	if err := readConfig(cfgFilePath, &cfg); err != nil {
		return err
	}
	if err := validateConfig(&cfg); err != nil {
		return errors.Wrap(err, "please fix the configuration file")
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
	updateAppWithConfig(app, &cfg, filepath.Dir(cfgFilePath))
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
			Name:  "config, c",
			Usage: "configuration file path",
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "configuration file name. The configuration file is searched from the current directory to the root directory recursively",
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
	help := ""
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
			cli.CommandHelpTemplate, "[arguments...]", strings.Join(argNames, " "), 1) + `ARGUMENTS:
` + strings.Join(argHelps, "\n")
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
		Action:             cliutil.WrapAction(newCommandAction(task, wd, scriptEnvs)),
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
			if arg.Prompt.Type == confirmPromptType {
				ans := false
				// TODO handle returned error
				// set the default value
				survey.AskOne(prompt, &ans)
				vars[arg.Name] = ans
				continue
			}
			ans := ""
			if err := survey.AskOne(prompt, &ans); err != nil {
				ans = arg.Default
			}
			vars[arg.Name] = ans
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

			p := createPrompt(flag.Prompt)
			if p != nil {
				if flag.Prompt.Type == confirmPromptType {
					ans := false
					// TODO handle returned error
					// set the default value
					survey.AskOne(p, &ans)
					vars[flag.Name] = ans
					continue
				}
				ans := ""
				if err := survey.AskOne(p, &ans); err != nil {
					ans = c.String(flag.Name)
				}
				vars[flag.Name] = ans
				continue
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
			v, ok := vars[k].(string)
			if !ok {
				return fmt.Errorf(
					"failed to convert the variable's value to the string: var: %s, value: %v", k, vars[k])
			}
			for _, e := range envNames {
				envs = append(envs, e+"="+v)
			}
		}

		for k, v := range task.Environment {
			envs = append(envs, k+"="+v)
		}

		scr, err := renderTemplate(task.Script, vars)
		if err != nil {
			return errors.Wrap(err, "failed to parse the script: "+task.Script)
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
		return "", errors.Wrap(err, "failed to get the current directory path")
	}
	for _, name := range names {
		p, err := cliutil.FindFile(wd, name, func(name string) bool {
			_, err := os.Stat(name)
			return err == nil
		})
		if err == nil {
			return p, nil
		}
	}
	return "", errors.New("the configuration file is not found")
}
