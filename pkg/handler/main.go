package handler

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/suzuki-shunsuke/cmdx/pkg/config"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	action "github.com/suzuki-shunsuke/cmdx/pkg/task-action"
	"github.com/suzuki-shunsuke/cmdx/pkg/util"
	"github.com/suzuki-shunsuke/cmdx/pkg/validate"
	"github.com/urfave/cli/v2"
)

const (
	boolFlagType   = "bool"
	defaultTimeout = 36000 // default 10H

	rootHelp = `# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/cmdx/refs/heads/main/json-schema/cmdx.json
cmdx - task runner https://github.com/suzuki-shunsuke/cmdx

Please run "cmdx help" to show help.
`

	appUsage = "task runner"
)

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (flags *LDFlags) AppVersion() string {
	return flags.Version + " (" + flags.Commit + ")"
}

func Main(flags *LDFlags, args []string) error {
	app := cli.NewApp()
	setupApp(app, flags)

	// Disable the builtin help command.
	//
	// If app.HideHelpCommand is false, help command doesn't work well because the default help command is run.
	// $ go run ./cmd/cmdx help
	// cmdx - task runner
	// https://github.com/suzuki-shunsuke/cmdx
	// Please run "cmdx help" to show help.
	//
	// If app.HideHelp is true, --help flag doesn't work well.
	// $ go run ./cmd/cmdx --help
	// Incorrect Usage: flag: help requested
	// flag: help requested
	//
	app.HideHelpCommand = true

	app.BashComplete = rootBashCompletion(flags, args)

	app.Action = mainAction(flags, args)

	app.CustomAppHelpTemplate = rootHelp
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return app.RunContext(ctx, args)
}

func mainAction(flags *LDFlags, args []string) func(*cli.Context) error {
	return func(c *cli.Context) error {
		cfg := domain.Config{}
		cfgFilePath := c.String("config")
		initFlag := c.Bool("init")
		listFlag := c.Bool("list")
		helpFlag := c.Bool("help")
		workingDirFlag := c.String("working-dir")
		cfgFileName := c.String("name")
		cfgClient := config.New()
		if initFlag {
			if cfgFilePath != "" {
				return cfgClient.Create(cfgFilePath)
			}
			if cfgFileName != "" {
				return cfgClient.Create(cfgFileName)
			}
			return cfgClient.Create(".cmdx.yaml")
		}

		if cfgFilePath == "" {
			var err error
			cfgFilePath, err = cfgClient.GetFilePath(cfgFileName)
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

		if err := cfgClient.Read(cfgFilePath, &cfg); err != nil {
			return err
		}
		if err := validate.Config(&cfg); err != nil {
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
		setupApp(app, flags)
		if workingDirFlag == "" {
			workingDirFlag = filepath.Dir(cfgFilePath)
		}
		var quiet *bool
		if c.IsSet("quiet") {
			q := c.Bool("quiet")
			quiet = &q
		}
		updateAppWithConfig(app, &cfg, &domain.GlobalFlags{
			DryRun:     c.Bool("dry-run"),
			Quiet:      quiet,
			WorkingDir: workingDirFlag,
		})
		return app.RunContext(c.Context, args)
	}
}

func setupApp(app *cli.App, flags *LDFlags) {
	app.Name = "cmdx"
	app.Version = flags.AppVersion()
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

func newFlag(flag domain.Flag) cli.Flag {
	switch flag.Type {
	case boolFlagType:
		f := &cli.BoolFlag{
			Name:    flag.Name,
			Usage:   flag.Usage,
			EnvVars: flag.InputEnvs,
		}
		if flag.Short != "" {
			f.Aliases = []string{flag.Short}
		}
		return f
	default:
		f := &cli.StringFlag{
			Name:    flag.Name,
			Usage:   flag.Usage,
			Value:   flag.Default,
			EnvVars: flag.InputEnvs,
		}
		if flag.Short != "" {
			f.Aliases = []string{flag.Short}
		}
		return f
	}
}

func getHelp(txt string, task domain.Task) string {
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

func convertTaskToCommand(task domain.Task, gFlags *domain.GlobalFlags) *cli.Command {
	help := getHelp(cli.CommandHelpTemplate, task)
	if !strings.HasSuffix(help, "\n") {
		help += "\n"
	}

	if len(task.Tasks) != 0 {
		tasks := make([]*cli.Command, len(task.Tasks))
		for i, s := range task.Tasks {
			tasks[i] = convertTaskToCommand(s, gFlags)
		}
		aliases := []string{}
		if task.Short != "" {
			aliases = []string{task.Short}
		}
		return &cli.Command{
			Name:               task.Name,
			Aliases:            aliases,
			Usage:              task.Usage,
			Description:        task.Description,
			Subcommands:        tasks,
			CustomHelpTemplate: help,
		}
	}

	flags := make([]cli.Flag, len(task.Flags))
	for j, flag := range task.Flags {
		flags[j] = newFlag(flag)
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

	var aliases []string
	if task.Short != "" {
		aliases = []string{task.Short}
	}

	return &cli.Command{
		Name:               task.Name,
		Aliases:            aliases,
		Usage:              task.Usage,
		Description:        task.Description,
		Flags:              flags,
		Action:             action.NewCommandAction(task, gFlags, scriptEnvs),
		CustomHelpTemplate: help,
	}
}

func updateAppWithConfig(app *cli.App, cfg *domain.Config, gFlags *domain.GlobalFlags) {
	cmds := make([]*cli.Command, len(cfg.Tasks))
	for i, task := range cfg.Tasks {
		cmds[i] = convertTaskToCommand(task, gFlags)
	}
	app.Commands = cmds
}

func setupEnvs(envs []string, name string) ([]string, error) {
	arr := make([]string, len(envs))
	for i, env := range envs {
		e, err := util.RenderTemplate(env, map[string]string{
			"name": name,
		})
		if err != nil {
			return nil, err
		}
		arr[i] = strings.ToUpper(strings.ReplaceAll(e, "-", "_"))
	}
	return arr, nil
}

func setupTask(task, base *domain.Task) error {
	inputEnvs := task.InputEnvs
	if len(inputEnvs) == 0 {
		inputEnvs = base.InputEnvs
	}

	if task.Quiet == nil {
		task.Quiet = base.Quiet
	}

	scriptEnvs := task.ScriptEnvs
	if len(scriptEnvs) == 0 {
		scriptEnvs = base.ScriptEnvs
	}

	if task.Environment == nil {
		task.Environment = map[string]string{}
	}
	for k, v := range base.Environment {
		if _, ok := task.Environment[k]; !ok {
			task.Environment[k] = v
		}
	}

	if task.Timeout.Duration == 0 {
		if base.Timeout.Duration == 0 {
			task.Timeout.Duration = defaultTimeout
		} else {
			task.Timeout.Duration = base.Timeout.Duration
		}
	}

	if task.Timeout.KillAfter == 0 {
		task.Timeout.KillAfter = base.Timeout.KillAfter
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

	task.Require.Exec = append(task.Require.Exec, base.Require.Exec...)
	task.Require.Environment = append(task.Require.Environment, base.Require.Environment...)

	if len(task.Tasks) != 0 {
		for i, t := range task.Tasks {
			if err := setupTask(&t, task); err != nil {
				return err
			}
			task.Tasks[i] = t
		}
		return nil
	}

	return nil
}

func setupConfig(cfg *domain.Config) error {
	base := &domain.Task{
		InputEnvs:   cfg.InputEnvs,
		Quiet:       cfg.Quiet,
		ScriptEnvs:  cfg.ScriptEnvs,
		Environment: cfg.Environment,
		Timeout:     cfg.Timeout,
	}
	for i, task := range cfg.Tasks {
		if err := setupTask(&task, base); err != nil {
			return err
		}
		cfg.Tasks[i] = task
	}
	return nil
}
