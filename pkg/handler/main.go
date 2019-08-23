package handler

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/suzuki-shunsuke/go-cliutil"
	"github.com/urfave/cli"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

const (
	boolFlagType = "bool"

	configurationFileTemplate = `---
# the configuration file of cmdx, which is a task runner.
# https://github.com/suzuki-shunsuke/cmdx
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
  #   bind_envs:
	#   - NAME
  environment:
    FOO: foo
  script: "echo $FOO"
`

	appUsage = "task runner"
)

type (
	Config struct {
		Tasks    []Task
		BindEnvs []string `yaml:"bind_envs"`
	}

	Task struct {
		Name        string
		Short       string
		Description string
		Usage       string
		Flags       []Flag
		Args        []Arg
		BindEnvs    []string `yaml:"bind_envs"`
		Environment map[string]string
		Script      string
	}

	Flag struct {
		Name     string
		Short    string
		Usage    string
		Default  string
		BindEnvs []string `yaml:"bind_envs"`
		Type     string
		Required bool
	}

	Arg struct {
		Name     string
		Usage    string
		Default  string
		BindEnvs []string `yaml:"bind_envs"`
		Required bool
	}
)

func Main() error {
	app := cli.NewApp()
	app.HideHelp = true
	setupApp(app)

	app.Action = cliutil.WrapAction(mainAction)

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
	env := strings.Join(flag.BindEnvs, ",")
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

	return cli.Command{
		Name:               task.Name,
		ShortName:          task.Short,
		Usage:              task.Usage,
		Description:        task.Description,
		Flags:              flags,
		Action:             cliutil.WrapAction(newCommandAction(task, wd)),
		CustomHelpTemplate: help,
	}
}

func updateVarsAndEnvsByArgs(args []Arg, cArgs []string, vars map[string]interface{}) ([]string, error) {
	n := len(cArgs)

	envs := []string{}
	for i, arg := range args {
		if i < n {
			val := cArgs[i]
			vars[arg.Name] = val
			for _, env := range arg.BindEnvs {
				envs = append(envs, env+"="+val)
			}
			continue
		}
		// the positional argument isn't given
		if arg.Default != "" {
			vars[arg.Name] = arg.Default
			for _, env := range arg.BindEnvs {
				envs = append(envs, env+"="+arg.Default)
			}
			continue
		}
		if arg.Required {
			return nil, fmt.Errorf("the %d th argument '%s' is required", i+1, arg.Name)
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
	return envs, nil
}

func newCommandAction(task Task, wd string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// create vars and envs
		// run command

		if err := validateFlagRequired(c, task.Flags); err != nil {
			return err
		}

		vars := map[string]interface{}{}

		envs, err := updateVarsAndEnvsByArgs(task.Args, c.Args(), vars)
		if err != nil {
			return err
		}
		envs = append(os.Environ(), envs...)

		for _, flag := range task.Flags {
			switch flag.Type {
			case boolFlagType:
				// don't ues c.Generic if flag.Type == "bool"
				// the value in the template is treated as false
				vars[flag.Name] = c.Bool(flag.Name)
			default:
				vars[flag.Name] = c.String(flag.Name)
			}
			for _, env := range flag.BindEnvs {
				envs = append(envs, env+"="+c.String(flag.Name))
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
			scr, wd, envs, c.GlobalBool("quiet"), c.GlobalBool("dry-run"))
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
			arr[i] = task.Name + " - " + task.Usage
		}
		fmt.Println(strings.Join(arr, "\n"))
		return nil
	}

	app := cli.NewApp()
	setupApp(app)
	updateAppWithConfig(app, &cfg, filepath.Dir(cfgFilePath))
	return app.Run(os.Args)
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

func setupConfig(cfg *Config) error {
	for i, task := range cfg.Tasks {
		if len(task.BindEnvs) == 0 && len(cfg.BindEnvs) != 0 {
			task.BindEnvs = cfg.BindEnvs
		}
		for j, flag := range task.Flags {
			if len(flag.BindEnvs) == 0 && len(task.BindEnvs) != 0 {
				flag.BindEnvs = task.BindEnvs
			}
			envs, err := setupEnvs(flag.BindEnvs, flag.Name)
			if err != nil {
				return err
			}
			flag.BindEnvs = envs
			task.Flags[j] = flag
		}

		for j, arg := range task.Args {
			if len(arg.BindEnvs) == 0 && len(task.BindEnvs) != 0 {
				arg.BindEnvs = task.BindEnvs
			}
			envs, err := setupEnvs(arg.BindEnvs, arg.Name)
			if err != nil {
				return err
			}
			arg.BindEnvs = envs
			task.Args[j] = arg
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
