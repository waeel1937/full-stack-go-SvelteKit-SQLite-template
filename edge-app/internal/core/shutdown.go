package core

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
}
