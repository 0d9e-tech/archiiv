package main

import "github.com/google/uuid"

func hasReadPerm(username string, uuid uuid.UUID) bool {
	// TODO
	// needs to fetch the meta data from fs and check that the user has
	// correct permission
	return true
}
