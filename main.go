package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Archiiv struct {
	fs    Fs
	files map[uuid.UUID]*File
	gin   *gin.Engine
}

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

func main() {
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		switch os.Args[1] {
		case "init":
			err := cmdInit()
			if err != nil {
				panic(err)
			}
		case "cli":
		}
	}

	dir := flag.String("dir", "/var/lib/archiiv", "Specify the Archív directory")
	flag.Parse()

	av := Archiiv{}

	rootFile, err := os.Open(filepath.Join(*dir, "fs", "root"))
	if err != nil {
		panic(err)
	}

	rootData, err := io.ReadAll(rootFile)
	if err != nil {
		panic(err)
	}

	rootUUID, err := uuid.ParseBytes(rootData)
	if err != nil {
		panic(err)
	}

	fs, err := NewFs(rootUUID, filepath.Join(*dir, "fs"))
	if err != nil {
		panic(err)
	}

	av.fs = fs
	err = av.loadFiles()
	if err != nil {
		panic(err)
	}

	av.gin = gin.Default()
	av.fsEndpoints()

	err = av.gin.Run()
	if err != nil {
		panic(err)
	}
}
