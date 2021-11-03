package action

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
	"github.com/suzuki-shunsuke/cmdx/pkg/execute"
	"github.com/suzuki-shunsuke/cmdx/pkg/flag"
	"github.com/suzuki-shunsuke/cmdx/pkg/prompt"
	"github.com/suzuki-shunsuke/cmdx/pkg/requirement"
	"github.com/suzuki-shunsuke/cmdx/pkg/util"
	"github.com/suzuki-shunsuke/cmdx/pkg/validate"
	"github.com/urfave/cli/v2"
)

func NewCommandAction(
	task domain.Task, gFlags *domain.GlobalFlags, scriptEnvs map[string][]string,
) cli.ActionFunc {
	return func(c *cli.Context) error {
		// create vars and envs
		// run command
		requireChecker := requirement.New()
		for _, requires := range task.Require.Exec {
			if err := requireChecker.Exec(requires); err != nil {
				return err
			}
		}
		for _, requires := range task.Require.Environment {
			if err := requireChecker.Env(requires); err != nil {
				return err
			}
		}

		vars := map[string]interface{}{}

		// get flag values and set them to vars
		if err := flag.SetValues(c, task.Flags, vars); err != nil {
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

		scr, err := util.RenderTemplate(task.Script, vars)
		if err != nil {
			return fmt.Errorf("failed to parse the script - %s: %w", task.Script, err)
		}

		quiet := false
		if gFlags.Quiet != nil {
			quiet = *gFlags.Quiet
		} else if task.Quiet != nil {
			quiet = *task.Quiet
		}

		if task.Type == "lambda" {
			return lambdaAction(c.Context, &task, vars)
		}

		exc := execute.New()

		return exc.Run(
			c.Context, &execute.Params{
				Shell:      task.Shell,
				Script:     scr,
				WorkingDir: gFlags.WorkingDir,
				Envs:       envs,
				Timeout: &execute.Timeout{
					Duration:  time.Duration(task.Timeout.Duration) * time.Second,
					KillAfter: time.Duration(task.Timeout.KillAfter) * time.Second,
				},
				Quiet:  quiet,
				DryRun: gFlags.DryRun,
			})
	}
}

func updateVarsByArgs(
	args []domain.Arg, cArgs []string, vars map[string]interface{},
) error {
	n := len(cArgs)

	for i, arg := range args {
		if i < n {
			val := cArgs[i]
			vars[arg.Name] = val
			if err := validate.ValueWithValidates(val, arg.Validate); err != nil {
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
				if err := validate.ValueWithValidates(v, arg.Validate); err != nil {
					return fmt.Errorf(arg.Name+" is invalid: %w", err)
				}
				break
			}
		}
		if isBoundEnv {
			continue
		}
		if prmpt := prompt.Create(arg.Prompt); prmpt != nil {
			val, err := prompt.GetValue(prmpt, arg.Prompt.Type)
			if err != nil {
				// TODO improvement
				if arg.Default != "" {
					vars[arg.Name] = arg.Default
					continue
				}
				continue
			}
			if v, ok := val.(string); ok {
				if err := validate.ValueWithValidates(v, arg.Validate); err != nil {
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
