package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Wait(timeout time.Duration, cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()
	cancel()
	<-ctx.Done()
}
