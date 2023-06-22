package main

import (
	"burp/internal/app"
	"burp/internal/server"
	"burp/pkg/shutdown"
	"context"
	"errors"
	"os"
	"time"
)

type ShutdownTask = func(ctx context.Context) error

func main() {
	app.Init()
	server.Init()
	<-shutdown.Shutdown(context.Background(), 5*time.Second, map[string]ShutdownTask{
		"cleanup_burp": func(ctx context.Context) error {
			return Cleanup()
		},
	})
	return
}

func Cleanup() error {
	if err := os.RemoveAll(".burp/"); err != nil {
		return errors.Join(errors.New("failed to cleanup .burp/ folder"), err)
	}
	return nil
}
