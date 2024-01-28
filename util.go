package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
)

type Error struct {
	Code int
	Msg  string
}

func newError(code int, msg string) Error {
	return Error{code, msg}
}

func (e Error) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Msg)
}

func respondError(e error, w http.ResponseWriter) {
	if aerr, ok := e.(Error); ok {
		w.WriteHeader(aerr.Code)
	} else {
		w.WriteHeader(500)
	}

	type resp struct {
		Ok  bool   `json:"ok"`
		Err string `json:"err"`
	}

	enc := json.NewEncoder(w)
	enc.Encode(&resp{false, e.Error()})
}

func respondOk(w http.ResponseWriter) {
	type resp struct {
		Ok bool `json:"ok"`
	}

	enc := json.NewEncoder(w)
	enc.Encode(&resp{true})
}

func toDataPath(sub, path string) string {
	return filepath.Join(g_cfg.DataDir, sub, path)
}

func toMetadataPath(path string) string {
	dir, file := filepath.Split(path)
	path = filepath.Join(dir, "__av_"+file+".json")

	return toDataPath("data", path)
}
