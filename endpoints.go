package main

import (
	"archiiv/fs"
	"archiiv/user"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func logAccesses(log *slog.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", "url", r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

func requireLogin(secret string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getSessionToken(r)
		if validateToken(secret, token) {
			h.ServeHTTP(w, r)
		} else {
			http.Error(w, "401 unauthorized", http.StatusUnauthorized)
		}
	})
}

func handleLogin(secret string, log *slog.Logger, userStore user.UserStore) http.Handler {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)

		var lr LoginRequest

		if err := dec.Decode(&lr); err != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("invalid json: %w", err))
			return
		}

		name := r.FormValue("name")
		pwd := r.FormValue("pwd")

		ok, token := login(name, pwd, userStore)

		if ok {
			log.Info("New login", "user", name)
			encodeOK(w, map[string]any{"token": token})
			return
		} else {
			log.Info("Failed login", "user", name)
			encodeError(w, http.StatusForbidden, errors.New("wrong name or password"))
			return
		}
	})
}

func handleWhoami() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := getUsername(r)
		encodeOK(w, struct {
			Name string `json:"name"`
		}{Name: name})
	})
}

func handleLs(fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")

		id, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		ch, e := fs.GetChildren(id)
		if e != nil {
			encodeError(w, http.StatusNotFound, fmt.Errorf("file not found: %w", e))
			return
		}

		// TODO(matěj) check permission

		encodeOK(w, ch)
	})
}

func handleCat(fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		sectionArg := r.PathValue("section")

		id, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		sectionReader, e := fs.OpenSection(id, sectionArg)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("open section: %w", e))
			return
		}

		if _, e = io.Copy(w, sectionReader); e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("io copy: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}

func handleUpload(fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		sectionArg := r.PathValue("section")

		uuid, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		sectionWriter, e := fs.CreateSection(uuid, sectionArg)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("create section: %w", e))
			return
		}

		if _, e = io.Copy(sectionWriter, r.Body); e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("io copy: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}

func handleTouch(fs *fs.Fs) http.Handler {
	type OkResponse struct {
		NewFileUuid uuid.UUID `json:"new_file_uuid"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		name := r.PathValue("name")

		parentId, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		fileId, e := fs.Touch(parentId, name)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("touch: %w", e))
			return
		}

		encodeOK(w, OkResponse{NewFileUuid: fileId})
	})
}

func handleMkdir(fs *fs.Fs) http.Handler {
	type OkResponse struct {
		NewDirUuid uuid.UUID `json:"new_dir_uuid"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		name := r.PathValue("name")

		id, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		fileId, e := fs.Mkdir(id, name)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("mkdir: %w", e))
			return
		}

		encodeOK(w, OkResponse{NewDirUuid: fileId})
	})
}

func handleMount(fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentArg := r.PathValue("parentUuid")
		childArg := r.PathValue("childUuid")

		parentUuid, e := uuid.Parse(parentArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		childUuid, e := uuid.Parse(childArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		e = fs.Mount(parentUuid, childUuid)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("parse uuid: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}

func handleUnmount(fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentArg := r.PathValue("parentUuid")
		childArg := r.PathValue("childUuid")

		parentUuid, e := uuid.Parse(parentArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		childUuid, e := uuid.Parse(childArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		e = fs.Unmount(parentUuid, childUuid)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("parse uuid: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}
