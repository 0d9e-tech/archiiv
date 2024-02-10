package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Archiiv struct {
	fs    Fs
	files map[uuid.UUID]*File
	gin   *gin.Engine
}

func main() {
	root, _ := uuid.Parse("a3762628-2da5-4c0d-80d5-5c7153b67321")

	av := Archiiv{}

	fs, err := NewFs(root, "test_fs")
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
