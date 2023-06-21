package main

import (
	"burp/burper/functions"
	"burp/reader"
	"burp/services"
	"context"
	"errors"
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ShutdownTask = func(ctx context.Context) error

func main() {
	functions.RegisterFunctions()
	tree, err := reader.Read("burp.toml")
	if err != nil {
		log.Fatalln(err)
		return
	}
	var burp services.Burp
	_, err = toml.Decode(tree.String(), &burp)
	if err != nil {
		log.Fatalln(err)
		return
	}
	dir, err := burp.Service.Clone()
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = burp.Environment.Save(*dir)
	if err != nil {
		log.Fatalln(err)
		return
	}
	<-Shutdown(context.Background(), 5*time.Second, map[string]ShutdownTask{
		"cleanup_burp": func(ctx context.Context) error {
			return Cleanup()
		},
	})
}

func Shutdown(ctx context.Context, timeout time.Duration, operations map[string]ShutdownTask) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-s

		log.Println("Preparing to shutdown burp-agent...")
		out := time.AfterFunc(timeout, func() {
			log.Fatalln("Graceful shutdown couldn't be completed even after timeout, forcing...")
		})
		defer out.Stop()
		var wg sync.WaitGroup
		for key, operation := range operations {
			wg.Add(1)

			key := key
			operation := operation
			go func() {
				defer wg.Done()
				if err := operation(ctx); err != nil {
					log.Println("Failed to complete shutdown task (", key, "): ", err)
					return
				}
				log.Println("Shutdown task (", key, ") was completed successfully.")
			}()
		}
		wg.Wait()
		close(wait)
	}()
	return wait
}

func Cleanup() error {
	if err := os.RemoveAll(".burp/"); err != nil {
		return errors.Join(errors.New("failed to cleanup .burp/ folder"), err)
	}
	return nil
}
