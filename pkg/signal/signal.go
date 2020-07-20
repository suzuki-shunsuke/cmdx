package signal

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var once sync.Once //nolint:gochecknoglobals

// Handle calls the function "callback" when the sinal is sent.
// This is useful to support canceling by signal.
// Usage:
//   c, cancel := context.WithCancel(ctx)
//   defer cancel()
//   go signal.Set(cancel)
//   ...
func Handle(callback func()) {
	once.Do(func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(
			signalChan, syscall.SIGHUP, syscall.SIGINT,
			syscall.SIGTERM, syscall.SIGQUIT)
		<-signalChan
		callback()
	})
}
