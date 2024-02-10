package main

import (
	"errors"

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
}
