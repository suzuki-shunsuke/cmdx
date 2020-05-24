package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

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
