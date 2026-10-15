package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UUIDPathParam parses a UUID path parameter and writes a consistent client
// error when the value is invalid.
func UUIDPathParam(c *gin.Context, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(c.Param(name))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":      "invalid_path_parameter",
			"message":   fmt.Sprintf("%s must be a valid UUID", name),
			"parameter": name,
		})
		return uuid.Nil, false
	}

	return value, true
}
