package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func handleLs(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")

		id, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		ch, e := fileStore.getChildren(id)
		if e != nil {
			encodeError(w, http.StatusNotFound, fmt.Errorf("file not found: %w", e))
			return
		}

		// TODO check permission

		encodeOK(w, ch)
	})
}

func handleCat(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		sectionArg := r.PathValue("section")

		id, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO check permission

		sectionReader, e := fileStore.openSection(id, sectionArg)
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

func handleUpload(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		sectionArg := r.PathValue("section")

		uuid, e := uuid.Parse(uuidArg)
		if e != nil {
			encodeError(w, http.StatusBadRequest, fmt.Errorf("parse uuid: %w", e))
			return
		}

		// TODO check permission

		sectionWriter, e := fileStore.createSection(uuid, sectionArg)
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

func handleTouch(fileStore fileStorer) http.Handler {
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

		// TODO check permission

		fileId, e := fileStore.touch(parentId, name)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("touch: %w", e))
			return
		}

		encodeOK(w, OkResponse{NewFileUuid: fileId})
	})
}

func handleMkdir(fileStore fileStorer) http.Handler {
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

		// TODO check permission

		fileId, e := fileStore.mkdir(id, name)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("mkdir: %w", e))
			return
		}

		encodeOK(w, OkResponse{NewDirUuid: fileId})
	})
}

func handleMount(fileStore fileStorer) http.Handler {
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

		// TODO check permission

		e = fileStore.mount(parentUuid, childUuid)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("parse uuid: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}

func handleUnmount(fileStore fileStorer) http.Handler {
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

		// TODO check permission

		e = fileStore.unmount(parentUuid, childUuid)
		if e != nil {
			encodeError(w, http.StatusInternalServerError, fmt.Errorf("parse uuid: %w", e))
			return
		}

		encodeOK[interface{}](w, nil)
	})
}
