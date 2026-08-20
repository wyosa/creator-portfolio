package store

import (
	"context"
	"path/filepath"
	"testing"

	"api/internal/models"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func createMedia(t *testing.T, st *Store, m *models.Media) *models.Media {
	t.Helper()
	created, err := st.CreateMedia(context.Background(), m)
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	return created
}

func TestCreateAndListOrdering(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	items, err := st.ListMedia(ctx)
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("expected empty non-nil slice, got %#v", items)
	}

	a := createMedia(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/a.jpg"})
	b := createMedia(t, st, &models.Media{Type: "video", Source: "youtube", ExternalID: "abc123"})
	c := createMedia(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/c.jpg"})

	if a.Position != 1 || b.Position != 2 || c.Position != 3 {
		t.Fatalf("expected positions 1,2,3 got %d,%d,%d", a.Position, b.Position, c.Position)
	}

	if err := st.Reorder(ctx, []int64{c.ID, a.ID, b.ID}); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	items, err = st.ListMedia(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	want := []int64{c.ID, a.ID, b.ID}
	for i, it := range items {
		if it.ID != want[i] {
			t.Fatalf("position %d: want id %d, got %d", i, want[i], it.ID)
		}
	}
}

func TestReorderValidation(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	a := createMedia(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/a.jpg"})
	b := createMedia(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/b.jpg"})

	cases := []struct {
		name string
		ids  []int64
	}{
		{"missing id", []int64{a.ID}},
		{"extra unknown id", []int64{a.ID, b.ID, 9999}},
		{"duplicate id", []int64{a.ID, a.ID}},
		{"only unknown ids", []int64{9998, 9999}},
	}
	for _, tc := range cases {
		if err := st.Reorder(ctx, tc.ids); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}

	// Failed reorders must not change positions.
	if err := st.Reorder(ctx, []int64{a.ID, b.ID}); err != nil {
		t.Fatalf("valid reorder: %v", err)
	}
	items, err := st.ListMedia(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if items[0].ID != a.ID || items[1].ID != b.ID {
		t.Fatalf("unexpected order after failed reorders: %d,%d", items[0].ID, items[1].ID)
	}
}

func TestPartialUpdate(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	m := createMedia(t, st, &models.Media{
		Type: "video", Source: "vimeo", ExternalID: "12345",
		Title: "old title", InstagramURL: "https://instagram.com/old",
	})

	newTitle := "new title"
	updated, err := st.UpdateMedia(ctx, m.ID, MediaPatch{Title: &newTitle})
	if err != nil {
		t.Fatalf("update title: %v", err)
	}
	if updated == nil {
		t.Fatal("updated is nil")
	}
	if updated.Title != "new title" {
		t.Fatalf("title not updated: %q", updated.Title)
	}
	if updated.InstagramURL != "https://instagram.com/old" {
		t.Fatalf("instagram_url should be untouched, got %q", updated.InstagramURL)
	}

	newURL := "https://instagram.com/new"
	updated, err = st.UpdateMedia(ctx, m.ID, MediaPatch{InstagramURL: &newURL})
	if err != nil {
		t.Fatalf("update instagram_url: %v", err)
	}
	if updated.InstagramURL != newURL {
		t.Fatalf("instagram_url not updated: %q", updated.InstagramURL)
	}
	if updated.Title != "new title" {
		t.Fatalf("title should be untouched, got %q", updated.Title)
	}

	// Missing id returns nil, nil.
	missing, err := st.UpdateMedia(ctx, 424242, MediaPatch{Title: &newTitle})
	if err != nil {
		t.Fatalf("update missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing id, got %#v", missing)
	}

	feat := true
	updated, err = st.UpdateMedia(ctx, m.ID, MediaPatch{Featured: &feat})
	if err != nil {
		t.Fatalf("update featured: %v", err)
	}
	if !updated.Featured {
		t.Fatal("featured not updated")
	}
	if updated.Title != "new title" {
		t.Fatalf("title should be untouched, got %q", updated.Title)
	}
}

func TestDelete(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	m := createMedia(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/x.jpg"})

	deleted, err := st.DeleteMedia(ctx, m.ID)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if deleted == nil || deleted.ID != m.ID {
		t.Fatalf("expected deleted row, got %#v", deleted)
	}

	// Second delete returns nil, nil.
	again, err := st.DeleteMedia(ctx, m.ID)
	if err != nil {
		t.Fatalf("delete missing: %v", err)
	}
	if again != nil {
		t.Fatalf("expected nil for missing id, got %#v", again)
	}

	items, err := st.ListMedia(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestTranslations(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	m := createMedia(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/t.jpg"})

	tr := map[string]map[string]string{
		"title":       {"en": "hello", "ru": "привет"},
		"description": {"en": "desc"},
	}
	if err := st.SetTranslations(ctx, "media", m.ID, tr); err != nil {
		t.Fatalf("set translations: %v", err)
	}

	loaded, err := st.LoadTranslations(ctx, "media", []int64{m.ID})
	if err != nil {
		t.Fatalf("load translations: %v", err)
	}
	if got := loaded[m.ID]["title"]["en"]; got != "hello" {
		t.Fatalf("title.en = %q, want hello", got)
	}
	if got := loaded[m.ID]["title"]["ru"]; got != "привет" {
		t.Fatalf("title.ru = %q, want привет", got)
	}
	if got := loaded[m.ID]["description"]["en"]; got != "desc" {
		t.Fatalf("description.en = %q, want desc", got)
	}

	// Upsert overwrites, empty value deletes the row.
	if err := st.SetTranslations(ctx, "media", m.ID, map[string]map[string]string{
		"title":       {"en": "hi", "ru": ""},
		"description": {"en": ""},
	}); err != nil {
		t.Fatalf("update translations: %v", err)
	}
	loaded, err = st.LoadTranslations(ctx, "media", []int64{m.ID})
	if err != nil {
		t.Fatalf("reload translations: %v", err)
	}
	if got := loaded[m.ID]["title"]["en"]; got != "hi" {
		t.Fatalf("title.en = %q, want hi", got)
	}
	if _, ok := loaded[m.ID]["title"]["ru"]; ok {
		t.Fatal("title.ru should be deleted after empty upsert")
	}
	if _, ok := loaded[m.ID]["description"]; ok {
		t.Fatal("description should be deleted after empty upsert")
	}

	// Unknown ids load nothing; empty id list is a no-op.
	loaded, err = st.LoadTranslations(ctx, "media", []int64{99999})
	if err != nil {
		t.Fatalf("load unknown id: %v", err)
	}
	if len(loaded) != 0 {
		t.Fatalf("expected empty map for unknown id, got %#v", loaded)
	}
	if _, err := st.LoadTranslations(ctx, "media", nil); err != nil {
		t.Fatalf("load empty ids: %v", err)
	}
}

func TestSettings(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	kv, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("get empty settings: %v", err)
	}
	if len(kv) != 0 {
		t.Fatalf("expected no settings, got %#v", kv)
	}

	if err := st.SetSettings(ctx, map[string]string{"a": "1", "b": "2"}); err != nil {
		t.Fatalf("set settings: %v", err)
	}
	// Upsert overwrites a single key, leaving the rest untouched.
	if err := st.SetSettings(ctx, map[string]string{"a": "3"}); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	kv, err = st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if kv["a"] != "3" || kv["b"] != "2" {
		t.Fatalf("settings = %#v, want a=3 b=2", kv)
	}
}

func TestSetSettingsAndTranslations(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	kv := map[string]string{"site_name": "Name"}
	tr := map[string]map[string]string{"site_name": {"ru": "Имя"}}
	if err := st.SetSettingsAndTranslations(ctx, kv, tr); err != nil {
		t.Fatalf("set settings and translations: %v", err)
	}

	settings, err := st.GetSettings(ctx)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings["site_name"] != "Name" {
		t.Fatalf("site_name = %q, want Name", settings["site_name"])
	}
	loaded, err := st.LoadTranslations(ctx, "settings", []int64{0})
	if err != nil {
		t.Fatalf("load translations: %v", err)
	}
	if got := loaded[0]["site_name"]["ru"]; got != "Имя" {
		t.Fatalf("settings translation = %q, want Имя", got)
	}
}

func TestReorderEdgeCases(t *testing.T) {
	st := openTestStore(t)
	ctx := context.Background()

	// Empty store: empty id list is a valid no-op.
	if err := st.Reorder(ctx, nil); err != nil {
		t.Fatalf("reorder empty store: %v", err)
	}

	m := createMedia(t, st, &models.Media{Type: "photo", Source: "upload", Path: "/media/solo.jpg", Title: "keep me"})
	if err := st.Reorder(ctx, []int64{m.ID}); err != nil {
		t.Fatalf("reorder single item: %v", err)
	}

	// Reorder must not clobber other fields.
	got, err := st.GetMedia(ctx, m.ID)
	if err != nil {
		t.Fatalf("get media: %v", err)
	}
	if got.Title != "keep me" {
		t.Fatalf("title = %q after reorder, want untouched", got.Title)
	}
	if got.Position != 0 {
		t.Fatalf("position = %d after single reorder, want 0", got.Position)
	}
}
