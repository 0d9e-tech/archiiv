package main

type File struct {
	Name  string           `json:"name"`
	Perms map[string]uint8 `json:"perms"`
	// TODO add the missing fields
}
