// Package user manipulates the users.json file and provides a simple API for
// the endpoints
package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type UserStore struct {
	// username to hashed password
	users map[string][64]byte
	// path of the users file
	path string
}

func (us UserStore) syncToDisk() error {
	file, err := os.Create(us.path)
	if err != nil {
		return err
	}

	err = json.NewEncoder(file).Encode(us.users)
	if err != nil {
		return err
	}

	return nil
}

func LoadUsers(path string) (us UserStore, err error) {
	us.path = filepath.Clean(path)

	usersFile, err := os.OpenFile(us.path, os.O_RDWR, 0)
	if err != nil {
		return
	}

	if err = json.NewDecoder(usersFile).Decode(&us.users); err != nil {
		err = fmt.Errorf("decode users file: %w", err)
		return
	}

	return
}

func (us UserStore) CheckPassword(name string, pwd [64]byte) bool {
	return us.users[name] == pwd
}

func (us UserStore) CreateUser(name string, pwd [64]byte) error {
	if _, ok := us.users[name]; ok {
		return errors.New("username already used")
	} else {
		us.users[name] = pwd

		err := us.syncToDisk()
		if err != nil {
			// undo the insert to keep the table consistent
			delete(us.users, name)
			return fmt.Errorf("createUser: %w", err)
		}

		return nil
	}
}

func (us UserStore) DeleteUser(name string) error {
	if _, ok := us.users[name]; !ok {
		return errors.New("deleting unknown user")
	}

	pwd := us.users[name]
	delete(us.users, name)

	err := us.syncToDisk()
	if err != nil {
		// undo the delete to keep the table consistent
		us.users[name] = pwd
		return fmt.Errorf("deleteUser: %w", err)
	}

	// TODO: GC user files here?

	return nil
}
