package handlers

import (
	"testing"
)

func TestUploadDiskPath(t *testing.T) {
	h := NewMediaHandler(nil, "/data/media")

	cases := []struct {
		name   string
		path   string
		want   string
		wantOK bool
	}{
		{"valid file", "/media/abc.jpg", "/data/media/abc.jpg", true},
		{"valid thumb file", "/media/abc.thumb.jpg", "/data/media/abc.thumb.jpg", true},
		{"empty path", "", "", false},
		{"no media prefix", "/etc/passwd", "", false},
		{"prefix without file", "/media/", "", false},
		{"prefix only", "/media", "", false},
		{"dot dot", "/media/..", "", false},
		{"dot", "/media/.", "", false},
		{"traversal", "/media/../secret", "", false},
		{"traversal deep", "/media/a/../../secret", "", false},
		{"subdirectory", "/media/a/b.jpg", "", false},
		{"absolute inside", "/media//etc/passwd", "", false},
		{"trailing slash", "/media/abc.jpg/", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := h.uploadDiskPath(tc.path)
			if ok != tc.wantOK {
				t.Fatalf("uploadDiskPath(%q) ok = %v, want %v", tc.path, ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("uploadDiskPath(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestValidateTranslations(t *testing.T) {
	cases := []struct {
		name    string
		tr      map[string]map[string]string
		fields  map[string]bool
		wantErr bool
	}{
		{"nil", nil, mediaTranslatableFields, false},
		{"empty", map[string]map[string]string{}, mediaTranslatableFields, false},
		{
			"valid media field",
			map[string]map[string]string{"title": {"en": "hello", "ru": "привет"}},
			mediaTranslatableFields,
			false,
		},
		{
			"all valid langs",
			map[string]map[string]string{"description": {"en": "a", "ru": "b", "es": "c", "et": "d", "de": "e"}},
			mediaTranslatableFields,
			false,
		},
		{
			"unknown media field",
			map[string]map[string]string{"slug": {"en": "x"}},
			mediaTranslatableFields,
			true,
		},
		{
			"unknown lang",
			map[string]map[string]string{"title": {"fr": "bonjour"}},
			mediaTranslatableFields,
			true,
		},
		{
			"valid settings field",
			map[string]map[string]string{"site_name": {"et": "nimi"}},
			settingsTranslatableFields,
			false,
		},
		{
			"media field rejected for settings",
			map[string]map[string]string{"title": {"en": "x"}},
			settingsTranslatableFields,
			true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTranslations(tc.tr, tc.fields)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateTranslations(%v) err = %v, wantErr %v", tc.tr, err, tc.wantErr)
			}
		})
	}
}
