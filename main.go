package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	DataDir string `json:"data_dir"`
	Debug   bool   `json:"debug"`
}

var (
	g_cfg Config = Config{
		Host:    "",
		Port:    4432,
		DataDir: "/var/lib/archiiv/",
		Debug:   false,
	}
)

func main() {
	cfgPath := flag.String("c", "/etc/archiiv/config.json", "Set the config.json path.")
	flag.Parse()

	f, err := os.Open(*cfgPath)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	dec := json.NewDecoder(f)

	err = dec.Decode(&g_cfg)
	if err != nil {
		slog.Error("Could not decode config:", "err", err)
		os.Exit(1)
	}

	registerAuthEndpoints()
	registerFsEndpoints()

	slog.Info(fmt.Sprintf("Listening on %s:%d", g_cfg.Host, g_cfg.Port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", g_cfg.Host, g_cfg.Port), nil))
}
