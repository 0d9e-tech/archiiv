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

func cmdInit() {
	os.Mkdir("fs", os.ModePerm)
	u := uuid.New()

	w, _ := os.Create("fs/root")
	w.WriteString(u.String())
	w.Close()

	w, _ = os.Create(filepath.Join("fs", u.String()))
	d, _ := json.Marshal(Record{
		Name:     "",
		Children: []uuid.UUID{},
	})
	w.Write(d)
	w.Close()

	w, _ = os.Create(filepath.Join("fs", u.String()+".meta"))
	d, _ = json.Marshal(File{
		UUID:      u,
		Type:      "archiiv/directory",
		Perms:     map[string]uint8{},
		Hooks:     []string{},
		CreatedBy: "root",
		CreatedAt: uint64(time.Now().Unix()),
	})
	w.Write(d)
	w.Close()
}

func main() {
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		switch os.Args[1] {
		case "init":
			cmdInit()
			return
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
	rootData, _ := io.ReadAll(rootFile)
	rootUUID, _ := uuid.ParseBytes(rootData)

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
