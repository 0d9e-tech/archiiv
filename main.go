package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	log.Info("Dobrý den")
	defer log.Info("Nashledanou")

	flags := flag.NewFlagSet("", flag.ContinueOnError)
	host := flags.String("host", "localhost", "")
	port := flags.String("port", "8275", "")
	secret := flags.String("secret", "hahahehe", "cryptographic secret") // TODO dont pass secrets as cli arguments
	err := flags.Parse(args)
	if err != nil {
		return err
	}

	users, err := loadUsers("users.json")
	if err != nil {
		return err
	}

	files, err := loadFiles()
	if err != nil {
		return err
	}

	srv := newServer(log, *secret, users, files)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(*host, *port),
		Handler: srv,
	}

	go func() {
		log.Info("listening", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listening and serving", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()

	return nil
}
