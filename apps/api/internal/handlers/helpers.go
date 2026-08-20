package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// fail writes a JSON error response.
func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

// failErr logs the underlying error and writes a JSON error response.
func failErr(c *gin.Context, status int, msg string, err error) {
	slog.Error(msg, "error", err)
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

// hasURLScheme reports whether raw parses as a URL using one of the allowed
// schemes (compared case-insensitively). It rejects scheme-relative and
// schemeless strings as well as dangerous schemes like javascript:.
func hasURLScheme(raw string, schemes ...string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	for _, s := range schemes {
		if strings.EqualFold(u.Scheme, s) {
			return true
		}
	}
	return false
}
