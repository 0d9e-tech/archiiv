package main

import (
	"archiiv/fs"
	"archiiv/user"
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
			encodeError(w, http.StatusUnauthorized, errors.New("401 unauthorized"))
		}
	})
}

func handleLogin(secret string, log *slog.Logger, userStore user.UserStore) http.Handler {
	type LoginRequest struct {
		Username string `json:"username"`
		Password [64]byte `json:"password"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lr, err := decode[LoginRequest](r)
		if err != nil {
			encodeError(w, http.StatusBadRequest, errors.New("wrong name or password"))
		}

		ok, token := login(lr.Username, lr.Password, secret, userStore)

		if ok {
			log.Info("New login", "user", lr.Username)
			encodeOK(w, struct {
				Token string `json:"token"`
			}{Token: token})
			return
		} else {
			log.Info("Failed login", "user", lr.Username)
			encodeError(w, http.StatusForbidden, errors.New("wrong name or password"))
			return
		}
	})
}

func handleWhoami(secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := getUsername(r, secret)
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

func handleUpload(log *slog.Logger, fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		sectionArg := r.PathValue("section")

		uuid, e := uuid.Parse(uuidArg)
		if e != nil {
			log.Error("handleUpload", "error", e)
			encodeError(w, http.StatusBadRequest, errors.New("invalid uuid"))
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
		NewFileUUID uuid.UUID `json:"new_file_uuid"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		name := r.PathValue("name")

		parentID, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		fileID, e := fs.Touch(parentID, name)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("touch: %w", e))
			return
		}

		encodeOK(w, OkResponse{NewFileUUID: fileID})
	})
}

func handleMkdir(fs *fs.Fs) http.Handler {
	type OkResponse struct {
		NewDirUUID uuid.UUID `json:"new_dir_uuid"`
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

		fileID, e := fs.Mkdir(id, name)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("mkdir: %w", e))
			return
		}

		encodeOK(w, OkResponse{NewDirUUID: fileID})
	})
}

func handleMount(fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentArg := r.PathValue("parentUUID")
		childArg := r.PathValue("childUUID")

		parentUUID, e := uuid.Parse(parentArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		childUUID, e := uuid.Parse(childArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		e = fs.Mount(parentUUID, childUUID)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("parse uuid: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}

func handleUnmount(fs *fs.Fs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentArg := r.PathValue("parentUUID")
		childArg := r.PathValue("childUUID")

		parentUUID, e := uuid.Parse(parentArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		childUUID, e := uuid.Parse(childArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO(matěj) check permission

		e = fs.Unmount(parentUUID, childUUID)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("parse uuid: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}
