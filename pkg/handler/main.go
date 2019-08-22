package handler

import (
	"bytes"
	"fmt"
	"os"
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
  #   env: NAME
  environment:
    FOO: foo
  script: "echo $FOO"
`

	appUsage = "task runner"
)

type (
	Config struct {
		Tasks []Task
	}

	Task struct {
		Name        string
		Short       string
		Description string
		Usage       string
		Flags       []Flag
		Args        []Arg
		Environment map[string]string
		Script      string
	}

	Flag struct {
		Name     string
		Short    string
		Usage    string
		Default  string
		Env      string
		Type     string
		Required bool
	}

	Arg struct {
		Name     string
		Usage    string
		Default  string
		Env      string
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
	}
}

func newFlag(flag Flag) cli.Flag {
	name := flag.Name
	if flag.Short != "" {
		name += ", " + flag.Short
	}
	switch flag.Type {
	case boolFlagType:
		return cli.BoolFlag{
			Name:   name,
			Usage:  flag.Usage,
			EnvVar: flag.Env,
		}
	default:
		return cli.StringFlag{
			Name:   name,
			Usage:  flag.Usage,
			Value:  flag.Default,
			EnvVar: flag.Env,
		}
	}
}

func convertTaskToCommand(task Task) cli.Command {
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
		Action:             cliutil.WrapAction(newCommandAction(task)),
		CustomHelpTemplate: help,
	}
}

func updateVarsAndEnvsByArgs(args []Arg, cArgs []string, envs []string, vars map[string]interface{}) ([]string, error) {
	n := len(cArgs)

	for i, arg := range args {
		if i < n {
			vars[arg.Name] = cArgs[i]
			if arg.Env != "" {
				envs = append(envs, arg.Env+"="+cArgs[i])
			}
			continue
		}
		// the positional argument isn't given
		if arg.Default != "" {
			vars[arg.Name] = arg.Default
			if arg.Env != "" {
				envs = append(envs, arg.Env+"="+arg.Default)
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

func newCommandAction(task Task) func(*cli.Context) error {
	return func(c *cli.Context) error {
		// create vars and envs
		// run command

		if err := validateFlagRequired(c, task.Flags); err != nil {
			return err
		}

		vars := map[string]interface{}{}

		envs, err := updateVarsAndEnvsByArgs(
			task.Args, c.Args(), os.Environ(), vars)
		if err != nil {
			return err
		}

		for _, flag := range task.Flags {
			switch flag.Type {
			case boolFlagType:
				// don't ues c.Generic if flag.Type == "bool"
				// the value in the template is treated as false
				vars[flag.Name] = c.Bool(flag.Name)
			default:
				vars[flag.Name] = c.String(flag.Name)
			}
			if flag.Env != "" {
				envs = append(envs, flag.Env+"="+c.String(flag.Name))
			}
		}

		for k, v := range task.Environment {
			envs = append(envs, k+"="+v)
		}

		scr, err := renderTemplate(task.Script, vars)
		if err != nil {
			return errors.Wrap(err, "failed to parse the script: "+task.Script)
		}

		return runScript(scr, envs, c.GlobalBool("quiet"))
	}
}

func updateAppWithConfig(app *cli.App, cfg *Config) {
	cmds := make([]cli.Command, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		cmds[i] = convertTaskToCommand(task)
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
	updateAppWithConfig(app, &cfg)
	return app.Run(os.Args)
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
