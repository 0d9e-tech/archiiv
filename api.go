package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func apiError(c *gin.Context, code int, err error) {
	slog.Error(err.Error())
	c.JSON(code, gin.H{
		"ok":    false,
		"error": err.Error(),
	})
}

func apiOk(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"ok":   true,
		"data": data,
	})
}
