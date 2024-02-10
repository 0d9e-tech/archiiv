package main

import (
	"errors"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (av *Archiiv) fsEndpoints() {
	group := av.gin.Group("/api/v1/fs", gin.BasicAuth(gin.Accounts{
		"marek": "123",
		"pub":   "pub",
		"root":  "supersecret",
	}))

	group.GET(":uuid/ls", func(c *gin.Context) {
		user, err := av.auth(c)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		uuid, err := uuid.Parse(c.Params.ByName("uuid"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		file, exists := av.files[uuid]
		if !exists {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		err = av.authFile(user, file, PermRead)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		apiOk(c, file.rec.Children)
	})

	group.GET(":uuid/section/:section", func(c *gin.Context) {
		user, err := av.auth(c)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		uuid, err := uuid.Parse(c.Params.ByName("uuid"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		file, exists := av.files[uuid]
		if !exists {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		err = av.authFile(user, file, PermRead)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		section := c.Params.ByName("section")

		f, err := file.rec.Open(section)
		if err != nil {
			apiError(c, 404, err)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			apiError(c, 500, err)
			return
		}

		var contentType string
		switch section {
		case "meta":
			contentType = "application/json"
		case "data":
			contentType = file.Type
		case "thumb":
			contentType = "image/webp"
		default:
			contentType = "application/octet-stream"
		}

		c.DataFromReader(200, stat.Size(), contentType, f, map[string]string{})
	})

	group.POST(":uuid/section/:section", func(c *gin.Context) {
		user, err := av.auth(c)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		uuid, err := uuid.Parse(c.Params.ByName("uuid"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		file, exists := av.files[uuid]
		if !exists {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		section := c.Params.ByName("section")

		p := PermWrite
		if section == "meta" {
			p |= PermOwner
		}

		err = av.authFile(user, file, p)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		f, err := file.rec.Create(section)
		if err != nil {
			apiError(c, 500, err)
			return
		}
		defer f.Close()

		_, err = io.Copy(f, c.Request.Body)
		if err != nil {
			apiError(c, 500, err)
			return
		}

		if section == "data" {
			file.Type = c.Request.Header.Get("Content-Type")
			err = file.Save()
			if err != nil {
				apiError(c, 500, err)
				return
			}
		}

		apiOk(c, nil)
	})

	group.POST(":uuid/touch/:name", func(c *gin.Context) {
		user, err := av.auth(c)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		uuid, err := uuid.Parse(c.Params.ByName("uuid"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		file, exists := av.files[uuid]
		if !exists {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		err = av.authFile(user, file, PermWrite)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		name := c.Params.ByName("name")
		fs := file.rec.fs
		rec, err := fs.NewRecord(name)
		if err != nil {
			apiError(c, 400, err)
			return
		}

		file.rec.Mount(rec)

		file = new(File)
		file.UUID = rec.UUID
		file.rec = rec
		file.CreatedAt = uint64(time.Now().Unix())
		file.CreatedBy = user
		file.Hooks = []string{}
		file.Type = "application/octet-stream"
		file.Perms = map[string]uint8{
			user: 0xff,
		}

		av.files[rec.UUID] = file

		err = file.Save()
		if err != nil {
			apiError(c, 500, err)
			return
		}

		apiOk(c, file.UUID)
	})

	group.POST(":uuid/mount/:uuid2", func(c *gin.Context) {
		user, err := av.auth(c)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		parent, err := uuid.Parse(c.Params.ByName("uuid"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		file, exists := av.files[parent]
		if !exists {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		err = av.authFile(user, file, PermWrite)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		uuid, err := uuid.Parse(c.Params.ByName("uuid2"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		parentRec := av.fs.GetRecord(parent)
		rec := av.fs.GetRecord(uuid)
		if rec == nil || parentRec == nil {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		parentRec.Mount(rec)
		apiOk(c, nil)
	})

	group.POST(":uuid/unmount/:uuid2", func(c *gin.Context) {
		user, err := av.auth(c)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		parent, err := uuid.Parse(c.Params.ByName("uuid"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		file, exists := av.files[parent]
		if !exists {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		err = av.authFile(user, file, PermWrite)
		if err != nil {
			apiError(c, 401, err)
			return
		}

		uuid, err := uuid.Parse(c.Params.ByName("uuid2"))
		if err != nil {
			apiError(c, 400, err)
			return
		}

		parentRec := av.fs.GetRecord(parent)
		rec := av.fs.GetRecord(uuid)
		if rec == nil || parentRec == nil {
			apiError(c, 404, errors.New("file not found"))
			return
		}

		parentRec.Unmount(rec)
		apiOk(c, nil)
	})
}
