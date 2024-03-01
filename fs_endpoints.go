package main

import (
	"io"
	"net/http"

	"github.com/google/uuid"
)

func handleLs(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")

		uuid, err := uuid.Parse(uuidArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, "invalid uuid")
			return
		}

		children, err := fileStore.getChildren(uuid)
		if err != nil {
			encodeError(w, http.StatusNotFound, "file not found")
			return
		}

		// TODO check permission

		encodeOK(w, children)
	})
}

func handleCat(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		sectionArg := r.PathValue("section")

		uuid, err := uuid.Parse(uuidArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// TODO check permission

		sectionReader, err := fileStore.openSection(uuid, sectionArg)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		_, err = io.Copy(w, sectionReader)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	})
}

func handleUpload(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		sectionArg := r.PathValue("section")

		uuid, err := uuid.Parse(uuidArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// TODO check permission

		sectionWriter, err := fileStore.createSection(uuid, sectionArg)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		_, err = io.Copy(sectionWriter, r.Body)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	})
}

func handleTouch(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		name := r.PathValue("name")

		uuid, err := uuid.Parse(uuidArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// TODO check permission

		fileStore.touch(uuid, name)
	})
}

func handleMkdir(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uuidArg := r.PathValue("uuid")
		name := r.PathValue("name")

		uuid, err := uuid.Parse(uuidArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// TODO check permission

		fileStore.mkdir(uuid, name)
	})
}

func handleMount(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentArg := r.PathValue("parentUuid")
		childArg := r.PathValue("childUuid")

		parentUuid, err := uuid.Parse(parentArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		childUuid, err := uuid.Parse(childArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// TODO check permission

		fileStore.mount(parentUuid, childUuid)
	})
}

func handleUnmount(fileStore fileStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentArg := r.PathValue("parentUuid")
		childArg := r.PathValue("childUuid")

		parentUuid, err := uuid.Parse(parentArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		childUuid, err := uuid.Parse(childArg)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		// TODO check permission

		fileStore.unmount(parentUuid, childUuid)
	})
}
