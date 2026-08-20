package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// fail writes a JSON error response.
func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

// failErr logs the underlying error and writes a JSON error response.
func failErr(c *gin.Context, status int, msg string, err error) {
	log.Printf("%s: %v", msg, err)
	fail(c, status, msg)
}

// parseID reads the :id route param, writing a 400 on failure.
func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
