package main

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"
)

type User struct {
	Name string
}

type PermType uint8

const (
	PermRead   = PermType(1 << 1)
	PermWrite  = PermType(1 << 2)
	PermMRead  = PermType(1 << 3)
	RermMWrite = PermType(1 << 4)
)

type Perms map[string]PermType

func getUserFromRequest(r *http.Request) (User, error) {
	return User{r.Header.Get("Authorization")}, nil
}

// TODO: this should be cached
func buildPerms(path string) (Perms, error) {
	// TODO: This doesn't seem right, but the filepath library doesn't have a function for this
	sp := strings.Split(path, "/")
	perms := Perms{}

	accum := ""
	for _, s := range sp {
		if accum != "" {
			accum = filepath.Join(accum, s)
		} else {
			accum = s
		}

		m, err := getMetadata(accum)
		if err != nil {
			return nil, err
		}

		for k, v := range m.Perms {
			perms[k] |= v
		}
	}

	return perms, nil
}

func (u *User) authPath(path string, pt PermType) error {
	if u.Name == "marian" {
		return errors.New("lol web dev")
	}

	perms, err := buildPerms(path)
	if err != nil {
		return err
	}

	if perms[u.Name]&pt != 0 || perms["pub"]&pt != 0 {
		return nil
	}

	return errors.New("Unauthorized")
}

func registerAuthEndpoints() {
	// TODO: change
	http.HandleFunc("/api/v1/auth/register", func(r http.ResponseWriter, w *http.Request) {
	})
}
