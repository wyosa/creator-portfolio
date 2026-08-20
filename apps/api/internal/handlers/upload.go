package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	mediaDir string
}

func NewUploadHandler(mediaDir string) *UploadHandler {
	return &UploadHandler{mediaDir: mediaDir}
}

var extKinds = map[string]string{
	".jpg":  "photo",
	".jpeg": "photo",
	".png":  "photo",
	".webp": "photo",
	".gif":  "photo",
	".mp4":  "video",
	".webm": "video",
	".mov":  "video",
	".m4v":  "video",
}

func (h *UploadHandler) Upload(c *gin.Context) {
	// Enforce 1 GiB body limit.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<30)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fail(c, http.StatusBadRequest, "missing or invalid file")
		return
	}
	defer func() { _ = file.Close() }()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	kind, ok := extKinds[ext]
	if !ok {
		fail(c, http.StatusBadRequest, "unsupported file extension")
		return
	}

	randBytes := make([]byte, 8)
	if _, err := rand.Read(randBytes); err != nil {
		failErr(c, http.StatusInternalServerError, "failed to generate filename", err)
		return
	}
	name := hex.EncodeToString(randBytes) + ext
	diskPath := filepath.Join(h.mediaDir, name)
	dst, err := os.Create(diskPath)
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to save file", err)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		_ = dst.Close()
		_ = os.Remove(diskPath)
		failErr(c, http.StatusInternalServerError, "failed to save file", err)
		return
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(diskPath)
		failErr(c, http.StatusInternalServerError, "failed to save file", err)
		return
	}

	// photos get a tiny blurred placeholder next to the original, and we
	// record pixel dimensions so the frontend can lay out justified grids
	thumb := ""
	width, height := 0, 0
	if kind == "photo" {
		img, err := decodeImage(diskPath)
		if err != nil {
			_ = os.Remove(diskPath)
			fail(c, http.StatusBadRequest, "invalid image file")
			return
		}
		bounds := img.Bounds()
		width, height = bounds.Dx(), bounds.Dy()
		if err := generateThumb(img, diskPath); err == nil {
			thumb = ThumbWebPath("/media/" + name)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"path": "/media/" + name, "kind": kind, "thumb": thumb,
		"width": width, "height": height,
	})
}
