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
)

const (
	configurationFileTemplate = `---
commands:
- name: hello
  # short: h
  description: hello command
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
)

type (
	Config struct {
		Commands []Command
	}

	Command struct {
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
		Binding  string
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
	setAppFlags(app)
	setAppCommands(app)

	app.Action = mainAction

	return app.Run(os.Args)
}

func updateAppWithConfig(app *cli.App, cfg *Config) {
	cmds := make([]cli.Command, len(cfg.Commands))
	for i, cmd := range cfg.Commands {
		flags := make([]cli.Flag, len(cmd.Flags))
		vars := map[string]interface{}{}
		for j, flag := range cmd.Flags {
			vars[flag.Name] = ""
			name := flag.Name
			if flag.Short != "" {
				name += ", " + flag.Short
			}
			switch flag.Type {
			case "bool":
				flags[j] = cli.BoolFlag{
					Name:     name,
					Usage:    flag.Usage,
					EnvVar:   flag.Env,
					Required: flag.Required,
				}
			default:
				flags[j] = cli.StringFlag{
					Name:     name,
					Usage:    flag.Usage,
					Value:    flag.Default,
					EnvVar:   flag.Env,
					Required: flag.Required,
				}
			}
		}

		cmds[i] = cli.Command{
			Name:        cmd.Name,
			ShortName:   cmd.Short,
			Usage:       cmd.Usage,
			Description: cmd.Description,
			Flags:       flags,
			Action: func(c *cli.Context) error {
				err := func() error {
					for _, flag := range cmd.Flags {
						vars[flag.Name] = c.Generic(flag.Name)
					}
					args := c.Args()
					n := c.NArg()
					for i, arg := range cmd.Args {
						if i >= n {
							if arg.Required {
								return fmt.Errorf("the %d th argument '%s' is required", i+1, arg.Name)
							}
							break
						}
						vars[arg.Name] = args[i]
					}
					extraArgs := []string{}
					for i, arg := range args {
						if i < len(cmd.Args) {
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
					scr, err := renderTemplate(cmd.Script, vars)
					if err != nil {
						return errors.Wrap(err, "failed to parse the script: "+cmd.Script)
					}

					command := exec.Command("sh", "-c", scr)
					command.Stdout = os.Stdout
					command.Stderr = os.Stderr
					envs := os.Environ()
					for k, v := range cmd.Environment {
						envs = append(envs, k+"="+v)
					}

					for _, flag := range cmd.Flags {
						if flag.Env != "" {
							envs = append(envs, flag.Env+"="+c.String(flag.Name))
						}
					}

					command.Env = append(os.Environ(), envs...)
					fmt.Println("+ " + scr)
					if err := command.Run(); err != nil {
						return err
					}
					return nil
				}()
				if _, ok := err.(*cli.ExitError); ok {
					return err
				}
				return cliutil.ConvErrToExitError(err)
			},
		}
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
			Usage: "configuration file name",
		},
		cli.BoolFlag{
			Name:  "init, i",
			Usage: "create the configuration file",
		},
	}
}

func setAppCommands(app *cli.App) {
	app.Commands = []cli.Command{
		{
			Name:      "help",
			Aliases:   []string{"h"},
			Usage:     "Shows a list of commands or help for one command",
			ArgsUsage: "[command]",
			Action: func(c *cli.Context) error {
				err := func() error {
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
					app := cli.NewApp()
					setAppFlags(app)
					setAppCommands(app)
					updateAppWithConfig(app, &cfg)
					return app.Run(os.Args)
				}()
				if _, ok := err.(*cli.ExitError); ok {
					return err
				}
				return cliutil.ConvErrToExitError(err)
			},
		},
	}
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
	app := cli.NewApp()
	setAppFlags(app)
	setAppCommands(app)
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
