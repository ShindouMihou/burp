package main

import (
	"burp/internal/app"
	"burp/internal/burper"
	"burp/internal/burpy"
	"burp/internal/services"
	"burp/pkg/shutdown"
	"context"
	"errors"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type ShutdownTask = func(ctx context.Context) error

func main() {
	app.Init()
	tree, err := burper.FromFile("burp.toml")
	if err != nil {
		log.Info().Err(err).Msg("Failed to parse TOML file into Burp tree")
		return
	}

	var burp services.Burp
	if err = toml.Unmarshal(tree.Bytes(), &burp); err != nil {
		log.Info().Err(err).Msg("Failed to parse TOML file into Burp services")
		return
	}

	err = burpy.Package(&burp)
	if err != nil {
		log.Info().Err(err).Msg("Failed to package files")
		return
	}
	// TODO: Add CLI application logic here.
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
