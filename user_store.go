package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type userStorer interface {
	checkPassword(name, pwd string) bool
	createUser(name, pwd string) error
	deleteUser(name string) error
}

type userStore struct {
	// username to hashed password
	users map[string]string
	file  *os.File
}

type user struct {
	Username string `json:"username"`
	Pwd      string `json:"password"`
}

func loadUsers(path string) (userStore, error) {
	usersFile, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}

	var us userStore

	us.file = usersFile

	err = json.NewDecoder(usersFile).Decode(&us.users)
	if err != nil {
		return us, fmt.Errorf("decode users file: %w", err)
	}

	return us, err
}

func (us userStore) syncToDisk() error {
	_, err := us.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = us.file.Truncate(0)
	if err != nil {
		return err
	}

	err = json.NewEncoder(us.file).Encode(us.users)
	if err != nil {
		return err
	}

	return nil
}

func (self userStore) checkPassword(name, pwd string) bool {
	return self.users[name] == pwd
}

func (us userStore) createUser(name, pwd string) error {
	if _, ok := us.users[name]; ok {
		return errors.New("username already used")
	} else {
		us.users[name] = pwd

		err := us.syncToDisk()
		if err != nil {
			delete(us.users, name)
			return fmt.Errorf("createUser: %w", err)
		}

		return nil
	}
}

func (us userStore) deleteUser(name string) error {
	if _, ok := us.users[name]; !ok {
		return errors.New("deleting unknown user")
	}

	pwd := us.users[name]
	delete(us.users, name)

	err := us.syncToDisk()
	if err != nil {
		us.users[name] = pwd
		return fmt.Errorf("deleteUser: %w", err)
	}

	// TODO: GC user files
	return nil
}
