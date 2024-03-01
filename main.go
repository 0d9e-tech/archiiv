package main

import (
	"archiiv/fs"
	"archiiv/user"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	srv, conf, err := createServer(log, os.Args[1:], os.Getenv)
	if err != nil {
		fmt.Printf("error from create: %s\n", err)
		os.Exit(1)
	}

	err = run(log, srv, conf)
	if err != nil {
		fmt.Printf("error from run: %s\n", err)
		os.Exit(1)
	}
}

func createServer(log *slog.Logger, args []string, env func(string) string) (http.Handler, config, error) {
	conf, err := getConfig(args, env)
	if err != nil {
		return nil, config{}, fmt.Errorf("get config: %w", err)
	}

	users, err := user.LoadUsers(conf.users_path)
	if err != nil {
		return nil, config{}, fmt.Errorf("load users: %w", err)
	}

	files, err := fs.NewFs(conf.root_uuid, conf.fs_root)
	if err != nil {
		return nil, config{}, fmt.Errorf("new fs: %w", err)
	}

	mux := http.NewServeMux()
	addRoutes(
		mux,
		log,
		conf.secret,
		users,
		files,
	)
	var srv http.Handler = mux
	srv = logAccesses(log, srv)

	return srv, conf, nil
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

	err = flags.Parse(args)
	if err != nil {
		err = fmt.Errorf("flags parse: %w", err)
		return
	}

	conf.root_uuid, err = uuid.Parse(root_uuid_string)

	if err != nil {
		err = fmt.Errorf("uuid parse: %w", err)
		return
	}

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

func run(log *slog.Logger, srv http.Handler, conf config) error {
	greet(log)
	defer goodbye(log)

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
