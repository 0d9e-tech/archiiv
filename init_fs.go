package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func cmdInit() error {
	os.Mkdir("fs", os.ModePerm)
	u := uuid.New()

	w, err := os.Create("fs/root")
	if err != nil {
		return err
	}

	w.WriteString(u.String())
	w.Close()

	w, err = os.Create(filepath.Join("fs", u.String()))
	if err != nil {
		return err
	}

	d, err := json.Marshal(Record{
		Name:     "",
		Children: []uuid.UUID{},
	})
	if err != nil {
		return err
	}

	w.Write(d)
	w.Close()

	w, err = os.Create(filepath.Join("fs", u.String()+".meta"))
	if err != nil {
		return err
	}

	d, err = json.Marshal(File{
		UUID:      u,
		Type:      "archiiv/directory",
		Perms:     map[string]uint8{},
		Hooks:     []string{},
		CreatedBy: "root",
		CreatedAt: uint64(time.Now().Unix()),
	})
	if err != nil {
		return err
	}

	w.Write(d)
	w.Close()

	return nil
}
