package main

import (
	"io"

	"github.com/google/uuid"
)

type fileStorer interface {
	getRoot() uuid.UUID

	getChildren(uuid.UUID) ([]uuid.UUID, error)

	mkdir(parent uuid.UUID, name string) (uuid.UUID, error)
	touch(parent uuid.UUID, name string) (uuid.UUID, error)

	mount(parent uuid.UUID, child uuid.UUID) error
	unmount(parent uuid.UUID, child uuid.UUID) error

	openSection(uuid uuid.UUID, section string) (io.ReadCloser, error)
	createSection(uuid uuid.UUID, section string) (io.WriteCloser, error)
	deleteSection(uuid uuid.UUID, section string) error
}
