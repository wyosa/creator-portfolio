package handlers

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestThumbPaths(t *testing.T) {
	cases := []struct {
		in   string
		disk string
		web  string
	}{
		{"/data/media/a.png", "/data/media/a.thumb.jpg", ""},
		{"/data/media/archive.tar.gz", "/data/media/archive.tar.thumb.jpg", ""},
		{"/media/a.png", "", "/media/a.thumb.jpg"},
		{"/media/noext", "", "/media/noext.thumb.jpg"},
	}
	for _, tc := range cases {
		if tc.disk != "" {
			if got := ThumbDiskPath(tc.in); got != tc.disk {
				t.Errorf("ThumbDiskPath(%q) = %q, want %q", tc.in, got, tc.disk)
			}
		}
		if tc.web != "" {
			if got := ThumbWebPath(tc.in); got != tc.web {
				t.Errorf("ThumbWebPath(%q) = %q, want %q", tc.in, got, tc.web)
			}
		}
	}
}

func TestGenerateThumbBounds(t *testing.T) {
	cases := []struct {
		name  string
		w, h  int
		wantW int
		wantH int
	}{
		{"tall narrow", 10, 2000, 1, 32},
		{"wide short", 2000, 10, 32, 1},
		{"landscape", 640, 480, 32, 24},
		{"portrait", 480, 640, 24, 32},
		{"small kept", 10, 20, 10, 20},
		{"square big", 100, 100, 32, 32},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := image.NewRGBA(image.Rect(0, 0, tc.w, tc.h))
			path := filepath.Join(t.TempDir(), "src.jpg")
			if err := generateThumb(src, path); err != nil {
				t.Fatalf("generateThumb: %v", err)
			}

			f, err := os.Open(ThumbDiskPath(path)) // #nosec G304 -- path built by the test
			if err != nil {
				t.Fatalf("open thumb: %v", err)
			}
			defer func() { _ = f.Close() }()
			cfg, _, err := image.DecodeConfig(f)
			if err != nil {
				t.Fatalf("decode thumb: %v", err)
			}
			if cfg.Width > 32 || cfg.Height > 32 {
				t.Fatalf("thumb = %dx%d, both sides must be <= 32", cfg.Width, cfg.Height)
			}
			if cfg.Width != tc.wantW || cfg.Height != tc.wantH {
				t.Errorf("thumb = %dx%d, want %dx%d", cfg.Width, cfg.Height, tc.wantW, tc.wantH)
			}
		})
	}
}
