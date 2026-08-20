package handlers

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"api/internal/models"
	"api/internal/store"

	"github.com/gin-gonic/gin"
)

func newMediaTestRouter(t *testing.T) (*gin.Engine, *store.Store, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	mediaDir := filepath.Join(dir, "media")
	if err := os.MkdirAll(mediaDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	st, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	h := NewMediaHandler(st, mediaDir)
	r := gin.New()
	r.GET("/media", h.List)
	r.POST("/media", h.Create)
	r.PUT("/media/:id", h.Update)
	r.DELETE("/media/:id", h.Delete)
	return r, st, mediaDir
}

func createMediaRow(t *testing.T, st *store.Store, m *models.Media) *models.Media {
	t.Helper()
	created, err := st.CreateMedia(context.Background(), m)
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	return created
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestMediaCreate(t *testing.T) {
	valid := gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg"}

	cases := []struct {
		name string
		body gin.H
		want int
	}{
		{"valid photo", valid, http.StatusCreated},
		{"valid youtube", gin.H{"type": "video", "source": "youtube", "external_id": "abc123"}, http.StatusCreated},
		{"valid dimensions", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "width": 100, "height": 100000}, http.StatusCreated},
		{"negative width", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "width": -1}, http.StatusBadRequest},
		{"huge width", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "width": 100001}, http.StatusBadRequest},
		{"huge height", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "height": 100001}, http.StatusBadRequest},
		{"javascript url", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "instagram_url": "javascript:alert(1)"}, http.StatusBadRequest},
		{"schemeless url", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "youtube_url": "example.com/x"}, http.StatusBadRequest},
		{"ftp url", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "vimeo_url": "ftp://example.com/x"}, http.StatusBadRequest},
		{"https url", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "instagram_url": "https://instagram.com/x"}, http.StatusCreated},
		{"http url", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "youtube_url": "http://youtube.com/x"}, http.StatusCreated},
		{"bad type", gin.H{"type": "gif", "source": "upload", "path": "/media/a.jpg"}, http.StatusBadRequest},
		{"photo from youtube", gin.H{"type": "photo", "source": "youtube", "external_id": "abc123"}, http.StatusBadRequest},
		{"youtube without id", gin.H{"type": "video", "source": "youtube"}, http.StatusBadRequest},
		{"upload without path", gin.H{"type": "photo", "source": "upload"}, http.StatusBadRequest},
		{"path traversal", gin.H{"type": "photo", "source": "upload", "path": "/media/../secret"}, http.StatusBadRequest},
		{"bad translation field", gin.H{"type": "photo", "source": "upload", "path": "/media/a.jpg", "translations": gin.H{"slug": gin.H{"en": "x"}}}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _, _ := newMediaTestRouter(t)
			w := doJSON(t, r, http.MethodPost, "/media", tc.body)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

func TestMediaCreateWithTranslations(t *testing.T) {
	r, _, _ := newMediaTestRouter(t)
	w := doJSON(t, r, http.MethodPost, "/media", gin.H{
		"type": "photo", "source": "upload", "path": "/media/a.jpg", "title": "base",
		"translations": gin.H{"title": gin.H{"ru": "заголовок"}},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
	var m models.Media
	decodeBody(t, w, &m)
	if got := m.Translations["title"]["ru"]; got != "заголовок" {
		t.Fatalf("translations.title.ru = %q, want заголовок", got)
	}
	if m.ID == 0 || m.Position != 1 {
		t.Errorf("id = %d, position = %d, want non-zero and 1", m.ID, m.Position)
	}
}

func TestMediaUpdate(t *testing.T) {
	r, st, mediaDir := newMediaTestRouter(t)
	m := createMediaRow(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/a.jpg"})

	t.Run("javascript url rejected", func(t *testing.T) {
		w := doJSON(t, r, http.MethodPut, "/media/"+itoa(m.ID), gin.H{"youtube_url": "javascript:alert(1)"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
		}
	})

	t.Run("valid url update", func(t *testing.T) {
		w := doJSON(t, r, http.MethodPut, "/media/"+itoa(m.ID), gin.H{"instagram_url": "https://instagram.com/new"})
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
		}
		var updated models.Media
		decodeBody(t, w, &updated)
		if updated.InstagramURL != "https://instagram.com/new" {
			t.Fatalf("instagram_url = %q", updated.InstagramURL)
		}
	})

	t.Run("not found", func(t *testing.T) {
		w := doJSON(t, r, http.MethodPut, "/media/4242", gin.H{"title": "x"})
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
	})

	t.Run("invalid preview path", func(t *testing.T) {
		w := doJSON(t, r, http.MethodPut, "/media/"+itoa(m.ID), gin.H{"preview_path": "/etc/passwd"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", w.Code)
		}
	})

	t.Run("replaced preview files are removed", func(t *testing.T) {
		withPreview := createMediaRow(t, st, &models.Media{
			Type: "photo", Source: "upload", Path: "/media/main.jpg", PreviewPath: "/media/old.jpg",
		})
		writeFile(t, filepath.Join(mediaDir, "main.jpg"))
		writeFile(t, filepath.Join(mediaDir, "old.jpg"))
		writeFile(t, filepath.Join(mediaDir, "old"+ThumbSuffix))

		w := doJSON(t, r, http.MethodPut, "/media/"+itoa(withPreview.ID), gin.H{"preview_path": "/media/new.jpg"})
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
		}
		if fileExists(filepath.Join(mediaDir, "old.jpg")) {
			t.Error("old preview file was not removed")
		}
		if fileExists(filepath.Join(mediaDir, "old"+ThumbSuffix)) {
			t.Error("old preview thumb was not removed")
		}
		if !fileExists(filepath.Join(mediaDir, "main.jpg")) {
			t.Error("main media file must stay")
		}
	})
}

func TestMediaDelete(t *testing.T) {
	r, st, mediaDir := newMediaTestRouter(t)

	m := createMediaRow(t, st, &models.Media{
		Type: "photo", Source: "upload", Path: "/media/a.jpg", PreviewPath: "/media/p.jpg",
	})
	for _, name := range []string{"a.jpg", "a" + ThumbSuffix, "p.jpg", "p" + ThumbSuffix} {
		writeFile(t, filepath.Join(mediaDir, name))
	}

	w := doJSON(t, r, http.MethodDelete, "/media/"+itoa(m.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	for _, name := range []string{"a.jpg", "a" + ThumbSuffix, "p.jpg", "p" + ThumbSuffix} {
		if fileExists(filepath.Join(mediaDir, name)) {
			t.Errorf("file %s was not removed", name)
		}
	}

	// Deleting again yields 404.
	w = doJSON(t, r, http.MethodDelete, "/media/"+itoa(m.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("second delete status = %d, want 404", w.Code)
	}
}
