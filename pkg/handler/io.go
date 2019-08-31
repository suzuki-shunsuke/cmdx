package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/suzuki-shunsuke/go-timeout/timeout"
)

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

func runScript(script, wd string, envs []string, tioCfg Timeout, quiet, dryRun bool) error {
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
		context.Background(), time.Duration(tioCfg.Duration)*time.Second)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT)
	resultChan := make(chan error, 1)
	defer cancel()
	sentSignals := map[os.Signal]struct{}{}
	go func() {
		resultChan <- runner.Run(ctx, cmd)
	}()
	var once sync.Once
	for {
		select {
		case <-ctx.Done():
			once.Do(func() {
				fmt.Fprintln(
					os.Stderr, "command timeout "+strconv.Itoa(tioCfg.Duration)+" seconds")
			})
		case err := <-resultChan:
			return err
		case sig := <-signalChan:
			if _, ok := sentSignals[sig]; ok {
				continue
			}
			sentSignals[sig] = struct{}{}
			fmt.Fprintf(os.Stderr, "send signal %d\n", sig)
			runner.SendSignal(sig.(syscall.Signal))
		}
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
