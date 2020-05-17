package handler

import (
	"fmt"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func rootBashCompletion(args []string) func(c *cli.Context) {
	return func(c *cli.Context) {
		cfg := Config{}
		cfgFilePath := c.String("config")
		initFlag := c.Bool("init")
		helpFlag := c.Bool("help")
		workingDirFlag := c.String("working-dir")
		cfgFileName := c.String("name")
		if initFlag {
			cli.DefaultAppComplete(c)
			return
		}

		if cfgFilePath == "" {
			var err error
			cfgFilePath, err = getConfigFilePath(cfgFileName)
			if err != nil {
				if helpFlag && cfgFileName == "" {
					cli.DefaultAppComplete(c)
					return
				}
				if c.NArg() == 1 && c.Args().First() == "help" && cfgFileName == "" {
					cli.DefaultAppComplete(c)
					return
				}
				if c.NArg() == 1 && c.Args().First() == "version" && cfgFileName == "" {
					cli.DefaultAppComplete(c)
					return
				}
				fmt.Println(err)
				return
			}
		}

		if err := readConfig(cfgFilePath, &cfg); err != nil {
			fmt.Println(err)
			return
		}
		if err := validateConfig(&cfg); err != nil {
			fmt.Println(fmt.Errorf("please fix the configuration file: %w", err))
			return
		}

		if err := setupConfig(&cfg); err != nil {
			fmt.Println(err)
			return
		}

		app := cli.NewApp()
		setupApp(app)
		if workingDirFlag == "" {
			workingDirFlag = filepath.Dir(cfgFilePath)
		}
		updateAppWithConfig(app, &cfg, &GlobalFlags{
			DryRun:     c.Bool("dry-run"),
			Quiet:      c.Bool("quiet"),
			WorkingDir: workingDirFlag,
		})
		if err := app.Run(args); err != nil {
			fmt.Println(err)
			return
		}
	}
}
