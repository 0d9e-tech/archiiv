package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	root, _ := uuid.Parse("a3762628-2da5-4c0d-80d5-5c7153b67321")
	fs, err := NewFs(root, "test_fs")
	if err != nil {
		panic(err)
	}

	rr := fs.GetRecord(root)

	/*folder, err := fs.NewRecord("marek")
	if err != nil {
		panic(err)
	}

	rr.Mount(folder)
	w, err := folder.Create("data")
	if err != nil {
		panic(err)
	}
	defer w.Close()

	w.WriteString("Amogus\n")*/

	for _, u := range rr.Children {
		rr.Unmount(fs.GetRecord(u))
	}

	fmt.Println(fs)
}
