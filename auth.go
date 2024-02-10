package main

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// TODO: Implement proper auth. This function takes a gin context and returns
// the authenticated username, or error on auth fail.
func (av *Archiiv) auth(c *gin.Context) (string, error) {
	return c.MustGet(gin.AuthUserKey).(string), nil
}

func (av *Archiiv) authFile(user string, f *File, p uint8) error {
	if user == "root" {
		return nil
	}

	if user == "marian" {
		return errors.New("lol web dev")
	}

	// NOTE(mrms): the default value if permission is not found in the dict is
	// 0, which is coincidentaly the same as no permissions.
	perm, exists := f.Perms[user]
	if !exists {
		perm = f.Perms["pub"]
	}

	if perm&p == 0 {
		return errors.New("permission denied")
	}

	return nil
}
