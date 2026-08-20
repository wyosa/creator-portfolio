package handlers

import (
	"net/http"
	"path/filepath"
	"testing"

	"api/internal/store"

	"github.com/gin-gonic/gin"
)

func newSettingsTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	h := NewSettingsHandler(st)
	r := gin.New()
	r.GET("/settings", h.Get)
	r.PUT("/settings", h.Update)
	return r
}

func getSettings(t *testing.T, r *gin.Engine) map[string]any {
	t.Helper()
	w := doJSON(t, r, http.MethodGet, "/settings", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /settings status = %d: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	decodeBody(t, w, &body)
	return body
}

func TestSettingsUpdateValidation(t *testing.T) {
	cases := []struct {
		name string
		body gin.H
	}{
		{"javascript link", gin.H{"info_links": []gin.H{{"label": "x", "url": "javascript:alert(1)", "enabled": true}}}},
		{"schemeless link", gin.H{"info_links": []gin.H{{"label": "x", "url": "example.com", "enabled": true}}}},
		{"empty link label", gin.H{"info_links": []gin.H{{"label": "", "url": "https://x.com", "enabled": true}}}},
		{"too many links", gin.H{"info_links": make([]gin.H, 21)}},
		{"unsupported language", gin.H{"languages": []string{"en", "fr"}}},
		{"empty languages", gin.H{"languages": []string{}}},
		{"bad translation field", gin.H{"translations": gin.H{"title": gin.H{"en": "x"}}}},
		{"bad translation lang", gin.H{"translations": gin.H{"site_name": gin.H{"fr": "x"}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newSettingsTestRouter(t)
			w := doJSON(t, r, http.MethodPut, "/settings", tc.body)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestSettingsRoundtrip(t *testing.T) {
	r := newSettingsTestRouter(t)

	w := doJSON(t, r, http.MethodPut, "/settings", gin.H{
		"site_name": "Dmitry",
		"languages": []string{"en", "ru"},
		"info_links": []gin.H{
			{"label": "email", "url": "mailto:hello@example.com", "enabled": true},
			{"label": "site", "url": "https://example.com", "enabled": false},
		},
		"translations": gin.H{"site_name": gin.H{"ru": "Дмитрий"}},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("PUT status = %d: %s", w.Code, w.Body.String())
	}

	body := getSettings(t, r)
	if body["site_name"] != "Dmitry" {
		t.Errorf("site_name = %v, want Dmitry", body["site_name"])
	}
	langs, ok := body["languages"].([]any)
	if !ok || len(langs) != 2 || langs[0] != "en" || langs[1] != "ru" {
		t.Errorf("languages = %#v, want [en ru]", body["languages"])
	}
	tr, ok := body["translations"].(map[string]any)
	if !ok {
		t.Fatalf("translations = %#v", body["translations"])
	}
	siteName, ok := tr["site_name"].(map[string]any)
	if !ok || siteName["ru"] != "Дмитрий" {
		t.Errorf("translations.site_name = %#v, want ru=Дмитрий", tr["site_name"])
	}
	links, ok := body["info_links"].([]any)
	if !ok || len(links) != 2 {
		t.Fatalf("info_links = %#v", body["info_links"])
	}
	if links[1].(map[string]any)["url"] != "https://example.com" {
		t.Errorf("second link = %#v", links[1])
	}

	// An empty translation value deletes it, falling back to the base text.
	w = doJSON(t, r, http.MethodPut, "/settings", gin.H{
		"translations": gin.H{"site_name": gin.H{"ru": ""}},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("PUT clear translation status = %d: %s", w.Code, w.Body.String())
	}
	body = getSettings(t, r)
	tr, ok = body["translations"].(map[string]any)
	if !ok || len(tr) != 0 {
		t.Errorf("translations after clear = %#v, want empty", body["translations"])
	}
	// The kv pair from before must survive a translations-only update.
	if body["site_name"] != "Dmitry" {
		t.Errorf("site_name after clear = %v, want Dmitry", body["site_name"])
	}
}

func TestSettingsFailedUpdateSavesNothing(t *testing.T) {
	r := newSettingsTestRouter(t)

	// Valid translations, invalid languages: the 400 must not persist anything.
	w := doJSON(t, r, http.MethodPut, "/settings", gin.H{
		"languages":    []string{"fr"},
		"translations": gin.H{"site_name": gin.H{"ru": "x"}},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	body := getSettings(t, r)
	tr, ok := body["translations"].(map[string]any)
	if !ok || len(tr) != 0 {
		t.Errorf("translations = %#v, want empty after failed update", body["translations"])
	}
}
