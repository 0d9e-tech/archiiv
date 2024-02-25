package main

import "github.com/google/uuid"

type fileStorer interface {
	getRoot() uuid.UUID

	getChildren(uuid.UUID) ([]uuid.UUID, error)

	mkdir(parent uuid.UUID, name string) (uuid.UUID, error)
	touch(parent uuid.UUID, name string) (uuid.UUID, error)
	delete(parent uuid.UUID, child uuid.UUID) error

	mount(parent uuid.UUID, child uuid.UUID) error

	readSection(uuid uuid.UUID, section string) ([]byte, error)

	// creates the section if it doesn't exist. overwrites previous data
	writeSection(uuid uuid.UUID, section string, data []byte) error
	deleteSection(uuid uuid.UUID, section string) error
}
