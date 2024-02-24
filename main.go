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
	if err := run(ctx, os.Stdout, os.Args, os.Getenv); err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

func newServer(
	log *slog.Logger,
	secret string,
	userStore userStorer,
	fileStore fileStorer,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		log,
		secret,
		userStore,
		fileStore,
	)
	var handler http.Handler = mux
	handler = logAccesses(log, handler)
	return handler
}

func getConfig(args []string, env func(string) string) (host string, port string, secret string, err error) {
	flags := flag.NewFlagSet("archiiv", flag.ContinueOnError)

	flags.StringVar(&host, "host", "localhost", "")
	flags.StringVar(&port, "port", "8275", "")

	secret = env("ARCHIIV_SECRET")

	err = flags.Parse(args[1:])

	return
}

func run(ctx context.Context, w io.Writer, args []string, env func(string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	log := slog.New(slog.NewJSONHandler(w, nil))
	log.Info("Dobrý den")
	defer log.Info("Nashledanou")

	host, port, secret, err := getConfig(args, env)
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

	srv := newServer(log, secret, users, files)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: srv,
	}

	go func() { // listening goroutine
		log.Info("listening", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listening and serving", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() { // cleanup goroutine
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("shutting down http server", "error", err)
		}
	}()
	wg.Wait()

	return nil
}
