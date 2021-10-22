package main

import (
	"fmt"
	"os"

	"github.com/suzuki-shunsuke/cmdx/pkg/handler"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	if err := handler.Main(&handler.LDFlags{
		Version: version,
		Commit:  commit,
		Date:    date,
	}, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ecerror.GetExitCode(err))
	}
}
