package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"api/internal/models"
	"api/internal/store"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	store    *store.Store
	mediaDir string
}

func NewMediaHandler(st *store.Store, mediaDir string) *MediaHandler {
	return &MediaHandler{store: st, mediaDir: mediaDir}
}

func (h *MediaHandler) List(c *gin.Context) {
	items, err := h.store.ListMedia(c.Request.Context())
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to list media", err)
		return
	}
	c.JSON(http.StatusOK, items)
}

var (
	youtubeIDRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	vimeoIDRe   = regexp.MustCompile(`^[0-9]+$`)
)

// ValidLangs are the languages the site supports for content translation.
var ValidLangs = map[string]bool{"en": true, "ru": true, "es": true, "et": true, "de": true}

var mediaTranslatableFields = map[string]bool{"title": true, "description": true}

// validateTranslations checks langs/fields against the allowed sets.
func validateTranslations(tr map[string]map[string]string, fields map[string]bool) error {
	for field, langs := range tr {
		if !fields[field] {
			return fmt.Errorf("field %q is not translatable", field)
		}
		for lang := range langs {
			if !ValidLangs[lang] {
				return fmt.Errorf("unsupported lang %q", lang)
			}
		}
	}
	return nil
}

type createMediaRequest struct {
	Type         string                       `json:"type"`
	Source       string                       `json:"source"`
	Path         string                       `json:"path"`
	ExternalID   string                       `json:"external_id"`
	Title        string                       `json:"title"`
	Description  string                       `json:"description"`
	InstagramURL string                       `json:"instagram_url"`
	YoutubeURL   string                       `json:"youtube_url"`
	VimeoURL     string                       `json:"vimeo_url"`
	Width        int                          `json:"width"`
	Height       int                          `json:"height"`
	Featured     bool                         `json:"featured"`
	Translations map[string]map[string]string `json:"translations"`
}

// maxDimension bounds client-supplied pixel dimensions.
const maxDimension = 100000

// validMediaURL allows an empty value (field unset) or an http/https URL.
func validMediaURL(u string) bool {
	return u == "" || hasURLScheme(strings.TrimSpace(u), "http", "https")
}

func (h *MediaHandler) Create(c *gin.Context) {
	var req createMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Type = strings.TrimSpace(req.Type)
	req.Source = strings.TrimSpace(req.Source)
	req.Path = strings.TrimSpace(req.Path)
	req.ExternalID = strings.TrimSpace(req.ExternalID)

	if req.Type != "photo" && req.Type != "video" {
		fail(c, http.StatusBadRequest, "type must be 'photo' or 'video'")
		return
	}
	// Photos must be uploads; videos may be upload/youtube/vimeo.
	if req.Type == "photo" && req.Source != "upload" {
		fail(c, http.StatusBadRequest, "photos must use source 'upload'")
		return
	}
	switch req.Source {
	case "upload":
		if req.Path == "" {
			fail(c, http.StatusBadRequest, "path is required for uploads")
			return
		}
		if _, ok := h.uploadDiskPath(req.Path); !ok {
			fail(c, http.StatusBadRequest, "invalid path")
			return
		}
	case "youtube":
		if req.ExternalID == "" {
			fail(c, http.StatusBadRequest, "external_id is required for youtube videos")
			return
		}
		if !youtubeIDRe.MatchString(req.ExternalID) {
			fail(c, http.StatusBadRequest, "invalid youtube id")
			return
		}
	case "vimeo":
		if req.ExternalID == "" {
			fail(c, http.StatusBadRequest, "external_id is required for vimeo videos")
			return
		}
		if !vimeoIDRe.MatchString(req.ExternalID) {
			fail(c, http.StatusBadRequest, "invalid vimeo id")
			return
		}
	default:
		fail(c, http.StatusBadRequest, "source must be 'upload', 'youtube' or 'vimeo'")
		return
	}
	if req.Width < 0 || req.Width > maxDimension || req.Height < 0 || req.Height > maxDimension {
		fail(c, http.StatusBadRequest, "width and height must be between 0 and 100000")
		return
	}
	for field, u := range map[string]string{
		"instagram_url": req.InstagramURL,
		"youtube_url":   req.YoutubeURL,
		"vimeo_url":     req.VimeoURL,
	} {
		if !validMediaURL(u) {
			fail(c, http.StatusBadRequest, field+" must be an http or https URL")
			return
		}
	}
	if req.Translations != nil {
		if err := validateTranslations(req.Translations, mediaTranslatableFields); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	m := &models.Media{
		Type:         req.Type,
		Source:       req.Source,
		Path:         req.Path,
		ExternalID:   req.ExternalID,
		Title:        req.Title,
		Description:  req.Description,
		InstagramURL: req.InstagramURL,
		YoutubeURL:   req.YoutubeURL,
		VimeoURL:     req.VimeoURL,
		Width:        req.Width,
		Height:       req.Height,
		Featured:     req.Featured,
	}
	created, err := h.store.CreateMedia(c.Request.Context(), m)
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to create media", err)
		return
	}
	if req.Translations != nil {
		if err := h.store.SetTranslations(c.Request.Context(), "media", created.ID, req.Translations); err != nil {
			failErr(c, http.StatusInternalServerError, "failed to save translations", err)
			return
		}
		created, err = h.store.GetMedia(c.Request.Context(), created.ID)
		if err != nil {
			failErr(c, http.StatusInternalServerError, "failed to reload media", err)
			return
		}
		if created == nil {
			fail(c, http.StatusInternalServerError, "failed to reload media")
			return
		}
	}
	c.JSON(http.StatusCreated, created)
}

type updateMediaRequest struct {
	Title        *string                      `json:"title"`
	Description  *string                      `json:"description"`
	InstagramURL *string                      `json:"instagram_url"`
	YoutubeURL   *string                      `json:"youtube_url"`
	VimeoURL     *string                      `json:"vimeo_url"`
	PreviewPath  *string                      `json:"preview_path"`
	Featured     *bool                        `json:"featured"`
	Translations map[string]map[string]string `json:"translations"`
}

func (h *MediaHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PreviewPath != nil {
		p := strings.TrimSpace(*req.PreviewPath)
		if p != "" {
			if _, ok := h.uploadDiskPath(p); !ok {
				fail(c, http.StatusBadRequest, "invalid preview_path")
				return
			}
		}
		req.PreviewPath = &p
	}
	for field, u := range map[string]*string{
		"instagram_url": req.InstagramURL,
		"youtube_url":   req.YoutubeURL,
		"vimeo_url":     req.VimeoURL,
	} {
		if u != nil && !validMediaURL(*u) {
			fail(c, http.StatusBadRequest, field+" must be an http or https URL")
			return
		}
	}
	if req.Translations != nil {
		if err := validateTranslations(req.Translations, mediaTranslatableFields); err != nil {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	existing, err := h.store.GetMedia(c.Request.Context(), id)
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to load media", err)
		return
	}
	if existing == nil {
		fail(c, http.StatusNotFound, "media not found")
		return
	}

	updated, err := h.store.UpdateMedia(c.Request.Context(), id, store.MediaPatch{
		Title:        req.Title,
		Description:  req.Description,
		InstagramURL: req.InstagramURL,
		YoutubeURL:   req.YoutubeURL,
		VimeoURL:     req.VimeoURL,
		PreviewPath:  req.PreviewPath,
		Featured:     req.Featured,
	})
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to update media", err)
		return
	}

	if req.Translations != nil {
		if err := h.store.SetTranslations(c.Request.Context(), "media", id, req.Translations); err != nil {
			failErr(c, http.StatusInternalServerError, "failed to save translations", err)
			return
		}
		updated, err = h.store.GetMedia(c.Request.Context(), id)
		if err != nil {
			failErr(c, http.StatusInternalServerError, "failed to reload media", err)
			return
		}
		if updated == nil {
			fail(c, http.StatusInternalServerError, "failed to reload media")
			return
		}
	}

	// A replaced preview leaves the old file behind; remove it unless it
	// doubles as the media file itself.
	if req.PreviewPath != nil && existing.PreviewPath != "" &&
		existing.PreviewPath != *req.PreviewPath && existing.PreviewPath != existing.Path {
		h.removeUpload(existing.PreviewPath)
	}
	c.JSON(http.StatusOK, updated)
}

func (h *MediaHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	deleted, err := h.store.DeleteMedia(c.Request.Context(), id)
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to delete media", err)
		return
	}
	if deleted == nil {
		fail(c, http.StatusNotFound, "media not found")
		return
	}
	if deleted.Source == "upload" && deleted.Path != "" {
		h.removeUpload(deleted.Path)
	}
	if deleted.PreviewPath != "" && deleted.PreviewPath != deleted.Path {
		h.removeUpload(deleted.PreviewPath)
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// removeUpload deletes an uploaded file along with the thumbnail generated
// next to it, if any. Missing files are fine, other failures are logged.
func (h *MediaHandler) removeUpload(webPath string) {
	diskPath, ok := h.uploadDiskPath(webPath)
	if !ok {
		return
	}
	for _, p := range []string{diskPath, ThumbDiskPath(diskPath)} {
		if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("failed to remove file", "path", p, "error", err)
		}
	}
}

// uploadDiskPath maps a stored path like /media/<name> to a file under
// mediaDir, guarding against path traversal.
func (h *MediaHandler) uploadDiskPath(p string) (string, bool) {
	name, ok := strings.CutPrefix(p, "/media/")
	if !ok || name == "" || filepath.Base(name) != name {
		return "", false
	}
	full := filepath.Join(h.mediaDir, name)
	if !strings.HasPrefix(full, filepath.Clean(h.mediaDir)+string(filepath.Separator)) {
		return "", false
	}
	return full, true
}

type reorderRequest struct {
	IDs []int64 `json:"ids"`
}

func (h *MediaHandler) Reorder(c *gin.Context) {
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.store.Reorder(c.Request.Context(), req.IDs); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
