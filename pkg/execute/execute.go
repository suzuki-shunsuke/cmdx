package execute

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

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

type Params struct {
	Shell      []string
	Script     string
	WorkingDir string
	Envs       []string
	Quiet      bool
	DryRun     bool
	Timeout    *Timeout
}

type Timeout struct {
	Duration  time.Duration
	KillAfter time.Duration
}

func (exc *Executor) Run(ctx context.Context, params *Params) error {
	shell := params.Shell
	if len(shell) == 0 {
		shell = []string{"sh", "-c"}
	}
	cmd := exec.Command(shell[0], append(shell[1:], params.Script)...) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = params.WorkingDir

	cmd.Env = append(os.Environ(), params.Envs...)
	if !params.Quiet {
		fmt.Fprintln(os.Stderr, "+ "+params.Script)
	}
	if params.DryRun {
		return nil
	}
	runner := timeout.NewRunner(params.Timeout.KillAfter)
	runner.SetSigKillCaballback(func(targetID int) {
		fmt.Fprintf(os.Stderr, "send SIGKILL to %d\n", targetID)
	})

	if params.Timeout.Duration > 0 {
		c, cancel := context.WithTimeout(ctx, params.Timeout.Duration)
		defer cancel()
		ctx = c
	}
	go func() {
		<-ctx.Done()
		err := ctx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintf(os.Stderr, "command is terminated by timeout: %d seconds\n", params.Timeout.Duration)
		}
	}()
	if err := runner.Run(ctx, cmd); err != nil {
		return ecerror.Wrap(err, cmd.ProcessState.ExitCode())
	}
	return nil
}
