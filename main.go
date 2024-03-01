package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	if err := run(os.Stdout, os.Args, os.Getenv); err != nil {
		fmt.Printf("error from main: %s\n", err)
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

type config struct {
	host       string
	port       string
	secret     string
	users_path string
	fs_root    string
	root_uuid  uuid.UUID
}

func getConfig(args []string, env func(string) string) (conf config, err error) {
	flags := flag.NewFlagSet("archiiv", flag.ContinueOnError)

	flags.StringVar(&conf.host, "host", "localhost", "")
	flags.StringVar(&conf.port, "port", "8275", "")
	flags.StringVar(&conf.fs_root, "fs_root", "", "")
	flags.StringVar(&conf.users_path, "users_path", "", "")

	var root_uuid_string string

	flags.StringVar(&root_uuid_string, "root_uuid", "", "")

	conf.secret = env("ARCHIIV_SECRET")

	err = flags.Parse(args[1:])
	if err != nil {
		return
	}

	conf.root_uuid, err = uuid.Parse(root_uuid_string)

	return
}

func greet(log *slog.Logger) {
	hour := time.Now().Hour()
	switch {
	case hour < 12:
		log.Info("Good morning")
	case hour < 17:
		log.Info("Good afternoon")
	default:
		log.Info("Good evening")
	}
}

func goodbye(log *slog.Logger) {
	log.Info("Goodbye")
}

func run(w io.Writer, args []string, env func(string) string) error {
	log := slog.New(slog.NewJSONHandler(w, nil))
	greet(log)
	defer goodbye(log)

	conf, err := getConfig(args, env)
	if err != nil {
		return err
	}

	users, err := loadUsers(conf.users_path)
	if err != nil {
		return err
	}

	files, err := newFs(conf.root_uuid, conf.fs_root)
	if err != nil {
		return err
	}

	srv := newServer(log, conf.secret, users, files)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(conf.host, conf.port),
		Handler: srv,
	}

	log.Info("listening", "address", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("listening and serving", "error", err)
		return err
	}

	return nil
}
