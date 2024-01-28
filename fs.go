package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (u *User) mkdir(path string) error {
	dir, _ := filepath.Split(path)
	if err := u.authPath(dir, PermWrite); err != nil {
		return err
	}

	os.Mkdir(toDataPath("data", path), 0770)
	(&Metadata{
		Perms: map[string]PermType{
			u.Name: 0xff,
		},
		CreatedBy: u.Name,
		Edits: []FileEdit{
			{
				User: u.Name,
				Date: uint64(time.Now().Unix()),
				Note: "Directory created",
			},
		},
	}).write(path)

	return nil
}

func (u *User) removeFile(path string) error {
	if err := u.authPath(path, PermWrite); err != nil {
		return err
	}

	if err := os.RemoveAll(toDataPath("data", path)); err != nil {
		return err
	}

	if err := os.Remove(toMetadataPath(path)); err != nil {
		return err
	}

	return nil
}

func registerFsEndpoints() {
	http.HandleFunc("/api/v1/fs/mkdir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			respondError(newError(405, "Method not allowed"), w)
			return
		}

		slog.Info("POST mkdir")
		u, err := getUserFromRequest(r)
		if err != nil {
			respondError(err, w)
			return
		}

		type Data struct {
			Path string `json:"path"`
		}

		d := Data{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err = dec.Decode(&d)
		if err != nil {
			respondError(err, w)
			return
		}

		err = u.mkdir(d.Path)
		if err != nil {
			respondError(err, w)
		}

		respondOk(w)
	})

	http.HandleFunc("/api/v1/fs/rm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			respondError(newError(405, "Method not allowed"), w)
			return
		}

		slog.Info("POST mkdir")
		u, err := getUserFromRequest(r)
		if err != nil {
			respondError(err, w)
			return
		}

		type Data struct {
			Path string `json:"path"`
		}

		d := Data{}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		err = dec.Decode(&d)
		if err != nil {
			respondError(err, w)
			return
		}

		err = u.removeFile(d.Path)
		if err != nil {
			respondError(err, w)
			return
		}

		respondOk(w)
	})
}
