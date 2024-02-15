package main

type userStorer interface {
	checkPassword(name, pwd string) bool
	createUser(name, pwd string) error
	deleteUser(name string) error
}

type userStore struct{}

func (self userStore) checkPassword(name, pwd string) bool {
	// TODO
	return true
}

func (self userStore) createUser(name, pwd string) error {
	// TODO
	return nil
}

func (self userStore) deleteUser(name, pwd string) error {
	// TODO
	return nil
}
