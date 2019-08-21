package handler

import (
	"fmt"
	"os"
	"os/exec"

	//	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/suzuki-shunsuke/go-cliutil"
	"github.com/urfave/cli"
)

const (
	helpTemplate = `
NAME:
   cmdx - task runner

USAGE:
   cmdx [global options] command [command options] [arguments...]

VERSION:
   0.1.0

AUTHOR:
   suzuki-shunsuke https://github.com/suzuki-shunsuke

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
`

	tasksHelpTemplate = `
init - create a configuration file if it doesn't exist
`

	taskHelpTemplate = `
init - create a configuration file if it doesn't exist

Positional arguments:

Options:
	--target, -t the target file path (required, default: "")

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
		Flags       []Flag
		Args        []Arg
		Environment map[string]string
		Script      string
	}

	Flag struct {
		Name        string
		Short       string
		Description string
		Default     string
		Binding     string
		Env         string
		Required    bool
	}

	Arg struct {
		Name        string
		Description string
		Default     string
		Env         string
		Required    bool
	}
)

func Main() error {
	// cmdx help -h --help
	// cmdx help hello
	// cmdx hello --help
	// cmdx -i --init
	// cmdx -c cmdx.yaml hello
	// app := kingpin.New("cmdx", "A command-line chat application.")
	// initFlag := app.Flag("init", "create the configuration file").Bool()
	// cfgFilePath := app.Flag("config", "the configuration file path").String()
	// cfgFileName := app.Flag("name", "the configuration file name").String()
	// helpCommand := app.Command("help", "Show this help")

	// cmdName := kingpin.MustParse(app.Parse(os.Args[1:]))
	// if *cfgFilePath != "" && *cfgFileName != "" {
	// 	return errors.New("the both --config and --name can't be used at the same time")
	// }

	app := cli.NewApp()
	app.HideHelp = true
	setAppFlags(app)
	setAppCommands(app)

	app.Action = mainAction

	return app.Run(os.Args)

	//	cfg := Config{}
	//	if *cfgFilePath != "" {
	//		app := cli.NewApp()
	//		cmds := make([]cli.Command, 0, len(cfg.Commands))
	//		for i, cmd := range cfg.Commands {
	//			flags := make([]cli.Flag, len(cmd.Flags))
	//			for i, flag := range cmd.Flags {
	//				flags[i] = cli.StringFlag{
	//					Name:   flag.Name,
	//					Usage:  flag.Description,
	//					Value:  flag.Default,
	//					EnvVar: flag.Env,
	//				}
	//			}
	//			cmds[i] = cli.Command{
	//				Name:        cmd.Name,
	//				ShortName:   cmd.Short,
	//				Description: cmd.Description,
	//				Flags:       flags,
	//			}
	//		}
	//		app.Commands = cmds
	//	}
	//	switch cmdName {
	//	case helpCommand.FullCommand():
	//	}

	// return nil
}

func updateAppWithConfig(app *cli.App, cfg *Config) {
	cmds := make([]cli.Command, len(cfg.Commands))
	for i, cmd := range cfg.Commands {
		flags := make([]cli.Flag, len(cmd.Flags))
		for j, flag := range cmd.Flags {
			flags[j] = cli.StringFlag{
				Name:     flag.Name,
				Usage:    flag.Description,
				Value:    flag.Default,
				EnvVar:   flag.Env,
				Required: flag.Required,
			}
		}
		cmds[i] = cli.Command{
			Name:        cmd.Name,
			ShortName:   cmd.Short,
			Description: cmd.Description,
			Flags:       flags,
			Action: func(c *cli.Context) error {
				command := exec.Command("sh", "-c", cmd.Script)
				command.Stdout = os.Stdout
				command.Stderr = os.Stderr
				envs := make([]string, len(cmd.Environment))
				i := 0
				for k, v := range cmd.Environment {
					envs[i] = k + "=" + v
					i++
				}
				command.Env = append(os.Environ(), envs...)
				fmt.Println("+ " + cmd.Script)
				if err := command.Run(); err != nil {
					return err
				}
				return nil
			},
		}
	}
	app.Commands = cmds
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
			},
		},
	}
}

func createConfigFile(p string) error {
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
