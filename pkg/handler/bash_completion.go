package handler

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/cmdx/pkg/config"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	"github.com/suzuki-shunsuke/cmdx/pkg/validate"
	"github.com/urfave/cli/v3"
)

func rootBashCompletion(flags *LDFlags, args []string) func(ctx context.Context, c *cli.Command) {
	return func(ctx context.Context, c *cli.Command) {
		cfg := domain.Config{}
		cfgFilePath := c.String("config")
		initFlag := c.Bool("init")
		helpFlag := c.Bool("help")
		cfgFileName := c.String("name")
		if initFlag {
			cli.DefaultAppComplete(ctx, c)
			return
		}

		cfgClient := config.New()

		if cfgFilePath == "" {
			var err error
			cfgFilePath, err = cfgClient.GetFilePath(cfgFileName)
			if err != nil {
				if helpFlag && cfgFileName == "" {
					cli.DefaultAppComplete(ctx, c)
					return
				}
				if c.NArg() == 1 && c.Args().First() == "help" && cfgFileName == "" {
					cli.DefaultAppComplete(ctx, c)
					return
				}
				if c.NArg() == 1 && c.Args().First() == "version" && cfgFileName == "" {
					cli.DefaultAppComplete(ctx, c)
					return
				}
				fmt.Println(err)
				return
			}
		}

		if err := cfgClient.Read(cfgFilePath, &cfg); err != nil {
			fmt.Println(err)
			return
		}
		if err := validate.Config(&cfg); err != nil {
			fmt.Println(fmt.Errorf("please fix the configuration file: %w", err))
			return
		}

		if err := setupConfig(&cfg); err != nil {
			fmt.Println(err)
			return
		}

		app := &cli.Command{}
		setupApp(app, flags)
		updateAppWithConfig(app, &cfg, &domain.GlobalFlags{})
		if err := app.Run(ctx, args); err != nil {
			fmt.Println(err)
			return
		}
	}
}
