package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/Songmu/timeout"
	"github.com/pkg/errors"
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
	tio := &timeout.Timeout{
		Cmd:       cmd,
		Duration:  tioCfg.Duration * time.Second,
		KillAfter: tioCfg.KillAfter * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan, syscall.SIGHUP, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGQUIT)
	resultChan := make(chan error, 1)
	defer cancel()
	go func() {
		err := func() error {
			status, err := tio.RunContext(ctx)
			if err != nil {
				return err
			}
			if status.IsKilled() {
				return errors.New("the command is killed")
			}
			if status.IsCanceled() {
				return errors.New("the command is canceled")
			}
			if status.IsTimedOut() {
				return fmt.Errorf("the command is timeout: %d sec", tioCfg.Duration)
			}
			return nil
		}()
		resultChan <- err
	}()
	for {
		select {
		case err := <-resultChan:
			return err
		case <-signalChan:
			cancel()
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
