package main

import (
	"fmt"
	"os"

	"github.com/suzuki-shunsuke/cmdx/pkg/handler"
	"github.com/suzuki-shunsuke/go-error-with-exit-code/ecerror"
)

func main() {
	if err := handler.Main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ecerror.GetExitCode(err))
	}
}
