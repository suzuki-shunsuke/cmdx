package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/suzuki-shunsuke/go-cliutil"
	"github.com/urfave/cli"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

const (
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
	setAppFlags(app)
	setAppCommands(app)
}

func newFlag(flag Flag) cli.Flag {
	name := flag.Name
	if flag.Short != "" {
		name += ", " + flag.Short
	}
	switch flag.Type {
	case "bool":
		return cli.BoolFlag{
			Name:     name,
			Usage:    flag.Usage,
			EnvVar:   flag.Env,
			Required: flag.Required,
		}
	default:
		return cli.StringFlag{
			Name:     name,
			Usage:    flag.Usage,
			Value:    flag.Default,
			EnvVar:   flag.Env,
			Required: flag.Required,
		}
	}
}

func newCommandWithConfig(task Task) cli.Command {
	flags := make([]cli.Flag, len(task.Flags))
	vars := map[string]interface{}{}
	for j, flag := range task.Flags {
		vars[flag.Name] = ""
		flags[j] = newFlag(flag)
	}

	return cli.Command{
		Name:        task.Name,
		ShortName:   task.Short,
		Usage:       task.Usage,
		Description: task.Description,
		Flags:       flags,
		Action:      cliutil.WrapAction(newCommandAction(task, vars)),
	}
}

func newCommandAction(task Task, vars map[string]interface{}) func(*cli.Context) error {
	return func(c *cli.Context) error {
		for _, flag := range task.Flags {
			vars[flag.Name] = c.Generic(flag.Name)
		}
		args := c.Args()
		n := c.NArg()
		envs := os.Environ()
		for i, arg := range task.Args {
			if i >= n {
				if arg.Default != "" {
					vars[arg.Name] = arg.Default
					if arg.Env != "" {
						envs = append(envs, arg.Env+"="+arg.Default)
					}
					continue
				}
				if arg.Required {
					return fmt.Errorf("the %d th argument '%s' is required", i+1, arg.Name)
				}
				continue
			}
			vars[arg.Name] = args[i]
			if arg.Env != "" {
				envs = append(envs, arg.Env+"="+args[i])
			}
		}
		extraArgs := []string{}
		for i, arg := range args {
			if i < len(task.Args) {
				continue
			}
			extraArgs = append(extraArgs, arg)
		}
		vars["_builtin"] = map[string]interface{}{
			"args":            extraArgs,
			"args_string":     strings.Join(extraArgs, " "),
			"all_args":        c.Args(),
			"all_args_string": strings.Join(c.Args(), " "),
		}
		scr, err := renderTemplate(task.Script, vars)
		if err != nil {
			return errors.Wrap(err, "failed to parse the script: "+task.Script)
		}

		command := exec.Command("sh", "-c", scr)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		for k, v := range task.Environment {
			envs = append(envs, k+"="+v)
		}

		for _, flag := range task.Flags {
			if flag.Env != "" {
				envs = append(envs, flag.Env+"="+c.String(flag.Name))
			}
		}

		command.Env = append(os.Environ(), envs...)
		if !c.GlobalBool("quiet") {
			fmt.Fprintln(os.Stderr, "+ "+scr)
		}
		if err := command.Run(); err != nil {
			return err
		}
		return nil
	}
}

func updateAppWithConfig(app *cli.App, cfg *Config) {
	cmds := make([]cli.Command, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		cmds[i] = newCommandWithConfig(task)
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

func readConfig(cfgFilePath string, cfg *Config) error {
	f, err := os.Open(cfgFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to open the configuration file: "+cfgFilePath)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return errors.Wrap(err, "failed to read the configuration file: "+cfgFilePath)
	}
	return nil
}

func setAppFlags(app *cli.App) {
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

func helpCommand(c *cli.Context) error {
	cfg := Config{}
	cfgFilePath := c.GlobalString("config")
	cfgFileName := c.GlobalString("name")
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
	app := cli.NewApp()
	setupApp(app)
	updateAppWithConfig(app, &cfg)
	return app.Run(os.Args)
}

func setAppCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:      "help",
			Aliases:   []string{"h"},
			Usage:     "show help",
			ArgsUsage: "[command]",
			Action:    cliutil.WrapAction(helpCommand),
		},
	}
}

func validateConfig(cfg *Config) error {
	taskNames := make(map[string]struct{}, len(cfg.Tasks))
	taskShortNames := make(map[string]struct{}, len(cfg.Tasks))
	for _, task := range cfg.Tasks {
		if _, ok := taskNames[task.Name]; ok {
			return errors.New(`the task name duplicates: "` + task.Name + `"`)
		}
		taskNames[task.Name] = struct{}{}

		if task.Short != "" {
			if _, ok := taskShortNames[task.Short]; ok {
				return errors.New(`the task short name duplicates: "` + task.Short + `"`)
			}
			taskShortNames[task.Short] = struct{}{}
		}

		flagNames := make(map[string]struct{}, len(task.Flags))
		flagShortNames := make(map[string]struct{}, len(task.Flags))
		for _, flag := range task.Flags {
			if len(flag.Short) > 1 {
				return fmt.Errorf(
					"The length of task.short should be 0 or 1. task: %s, flag: %s, short: %s",
					task.Name, flag.Name, flag.Short)
			}

			if _, ok := flagNames[flag.Name]; ok {
				return fmt.Errorf(
					`the flag name duplicates: task: "%s", flag: "%s"`,
					task.Name, flag.Name)
			}
			flagNames[flag.Name] = struct{}{}

			if flag.Short != "" {
				if _, ok := flagShortNames[flag.Short]; ok {
					return fmt.Errorf(
						`the flag short name duplicates: task: "%s", flag.short: "%s"`,
						task.Name, flag.Short)
				}
				flagShortNames[flag.Short] = struct{}{}
			}

			switch flag.Type {
			case "":
			case "bool":
			case "string":
			default:
				return fmt.Errorf(
					"The flag type should be either '' or 'string' or 'bool'. task: %s, flag: %s, flag.type: %s",
					task.Name, flag.Name, flag.Type)
			}
		}
		argNames := make(map[string]struct{}, len(task.Args))
		for _, arg := range task.Args {
			if _, ok := argNames[arg.Name]; ok {
				return fmt.Errorf(
					`the positional argument name duplicates: task: "%s", arg: "%s"`,
					task.Name, arg.Name)
			}
			argNames[arg.Name] = struct{}{}
		}
	}
	return nil
}

func createConfigFile(p string) error {
	if _, err := os.Stat(p); err == nil {
		// If the configuration file already exists, do nothing.
		return nil
	}
	if err := ioutil.WriteFile(p, []byte(configurationFileTemplate), 0644); err != nil {
		return errors.Wrap(err, "failed to create the configuration file: "+p)
	}
	return nil
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
