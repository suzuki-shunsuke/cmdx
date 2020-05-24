package handler

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

func readConfig(cfgFilePath string, cfg *Config) error {
	f, err := os.Open(cfgFilePath)
	if err != nil {
		return fmt.Errorf("failed to open the configuration file %s: %w", cfgFilePath, err)
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read the configuration file %s: %w", cfgFilePath, err)
	}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("failed to parse the configuration file. the configuration file is invalid: %s: %w", cfgFilePath, err)
	}
	return nil
}

func runScript(ctx context.Context, shell []string, script, wd string, envs []string, tioCfg Timeout, quiet, dryRun bool) error {
	if len(shell) == 0 {
		shell = []string{"sh", "-c"}
	}
	cmd := exec.Command(shell[0], append(shell[1:], script)...) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = wd

	cmd.Env = append(os.Environ(), envs...)
	if !quiet {
		fmt.Fprintln(os.Stderr, "+ "+script)
	}
	if dryRun {
		return nil
	}
	runner := timeout.NewRunner(time.Duration(tioCfg.KillAfter) * time.Second)
	runner.SetSigKillCaballback(func(targetID int) {
		fmt.Fprintf(os.Stderr, "send SIGKILL to %d\n", targetID)
	})

	ctx, cancel := context.WithTimeout(
		ctx, time.Duration(tioCfg.Duration)*time.Second)
	defer cancel()
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintf(os.Stderr, "command is terminated by timeout: %d seconds\n", tioCfg.Duration)
		}
	}()
	if err := runner.Run(ctx, cmd); err != nil {
		return ecerror.Wrap(err, cmd.ProcessState.ExitCode())
	}
	return nil
}

func createConfigFile(p string) error {
	if _, err := os.Stat(p); err == nil {
		// If the configuration file already exists, do nothing.
		return nil
	}
	if err := ioutil.WriteFile(p, []byte(configurationFileTemplate), 0644); err != nil { //nolint:gosec
		return fmt.Errorf("failed to create the configuration file %s: %w", p, err)
	}
	return nil
}
