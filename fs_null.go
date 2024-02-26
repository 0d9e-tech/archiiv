package main

import (
	"errors"
	"io"

	"github.com/google/uuid"
)

type NullFileStore struct{}

func (nfs NullFileStore) getRoot() uuid.UUID {
	return uuid.Nil
}

func (nfs NullFileStore) getChildren(uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{}, nil
}

func (nfs NullFileStore) mkdir(uuid.UUID, string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (nfs NullFileStore) touch(uuid.UUID, string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (nfs NullFileStore) unmount(parent uuid.UUID, child uuid.UUID) error {
	return nil
}

func (nfs NullFileStore) mount(parent uuid.UUID, newChild uuid.UUID) error {
	return nil
}

func (nfs NullFileStore) openSection(uuid uuid.UUID, section string) (io.ReadCloser, error) {
	return nil, errors.New("file not found")
}

func (nfs NullFileStore) createSection(uuid uuid.UUID, section string) (io.WriteCloser, error) {
	return nil, nil
}

func (nfs NullFileStore) deleteSection(uuid uuid.UUID, section string) error {
	return errors.New("file not found")
}
