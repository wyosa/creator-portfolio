package handlers

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newUploadTestRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mediaDir := t.TempDir()
	r := gin.New()
	r.POST("/upload", NewUploadHandler(mediaDir).Upload)
	return r, mediaDir
}

func uploadFile(t *testing.T, r *gin.Engine, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if filename != "" {
		part, err := mw.CreateFormFile("file", filename)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("write form file: %v", err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func pngBytes(t *testing.T, width, height int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// fakeMP4 builds an ftyp box the sniffer recognizes as video/mp4 (major brand
// isom plus an mp41 compatible brand, like ffmpeg produces).
func fakeMP4() []byte {
	body := []byte("ftypisom")
	body = append(body, make([]byte, 4)...)
	body = append(body, []byte("isomiso2avc1mp41")...)
	out := make([]byte, 4)
	// #nosec G115 -- the test builds a tiny ftyp box, no overflow possible.
	binary.BigEndian.PutUint32(out, uint32(4+len(body)))
	return append(out, body...)
}

// fakeMOV builds a QuickTime ftyp box; the sniffer does not know the "qt"
// brand and reports application/octet-stream.
func fakeMOV() []byte {
	body := []byte("ftypqt  ")
	body = append(body, make([]byte, 8)...)
	out := make([]byte, 4)
	// #nosec G115 -- the test builds a tiny ftyp box, no overflow possible.
	binary.BigEndian.PutUint32(out, uint32(4+len(body)))
	return append(out, body...)
}

func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestUploadValidation(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		content  []byte
	}{
		{"no file", "", nil},
		{"unsupported extension", "notes.txt", []byte("hello")},
		{"text as png", "fake.png", []byte("this is not an image")},
		{"html as mp4", "fake.mp4", []byte("<html><body>hi</body></html>")},
		{"html as mov", "fake.mov", []byte("<html><body>hi</body></html>")},
		{"text as webm", "fake.webm", []byte("not a video at all")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, mediaDir := newUploadTestRouter(t)
			w := uploadFile(t, r, tc.filename, tc.content)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
			}
			if names := dirEntries(t, mediaDir); len(names) != 0 {
				t.Fatalf("rejected upload left files behind: %v", names)
			}
		})
	}
}

func TestUploadPhoto(t *testing.T) {
	r, mediaDir := newUploadTestRouter(t)
	content := pngBytes(t, 640, 480)
	w := uploadFile(t, r, "photo.png", content)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Path   string `json:"path"`
		Kind   string `json:"kind"`
		Thumb  string `json:"thumb"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	decodeBody(t, w, &resp)
	if resp.Kind != "photo" {
		t.Errorf("kind = %q, want photo", resp.Kind)
	}
	if resp.Width != 640 || resp.Height != 480 {
		t.Errorf("dimensions = %dx%d, want 640x480", resp.Width, resp.Height)
	}
	if !strings.HasPrefix(resp.Path, "/media/") || !strings.HasSuffix(resp.Path, ".png") {
		t.Errorf("unexpected path %q", resp.Path)
	}
	wantThumb := strings.TrimSuffix(resp.Path, ".png") + ThumbSuffix
	if resp.Thumb != wantThumb {
		t.Errorf("thumb = %q, want %q", resp.Thumb, wantThumb)
	}

	// The original bytes and the thumbnail must both be on disk.
	name := strings.TrimPrefix(resp.Path, "/media/")
	// #nosec G304 -- path derived from the handler response inside t.TempDir().
	saved, err := os.ReadFile(filepath.Join(mediaDir, name))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if !bytes.Equal(saved, content) {
		t.Error("saved file content differs from uploaded bytes")
	}
	thumbPath := filepath.Join(mediaDir, strings.TrimSuffix(name, ".png")+ThumbSuffix)
	f, err := os.Open(thumbPath) // #nosec G304 -- path built by the test
	if err != nil {
		t.Fatalf("open thumb: %v", err)
	}
	defer func() { _ = f.Close() }()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decode thumb: %v", err)
	}
	if cfg.Width > 32 || cfg.Height > 32 {
		t.Errorf("thumb dimensions = %dx%d, want both <= 32", cfg.Width, cfg.Height)
	}
}

func TestUploadVideo(t *testing.T) {
	r, mediaDir := newUploadTestRouter(t)
	content := fakeMP4()
	w := uploadFile(t, r, "clip.mp4", content)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Path  string `json:"path"`
		Kind  string `json:"kind"`
		Thumb string `json:"thumb"`
	}
	decodeBody(t, w, &resp)
	if resp.Kind != "video" {
		t.Errorf("kind = %q, want video", resp.Kind)
	}
	if resp.Thumb != "" {
		t.Errorf("videos get no thumb, got %q", resp.Thumb)
	}

	// Sniffing must not eat the leading bytes: the saved file matches the
	// upload byte for byte.
	// #nosec G304 -- path derived from the handler response inside t.TempDir().
	saved, err := os.ReadFile(filepath.Join(mediaDir, strings.TrimPrefix(resp.Path, "/media/")))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if !bytes.Equal(saved, content) {
		t.Error("saved video content differs from uploaded bytes (reader not rewound?)")
	}
}

// QuickTime uploads sniff as application/octet-stream and must still pass
// (WHATWG sniffing has no "qt" brand), while text/html stays rejected.
func TestUploadMov(t *testing.T) {
	r, _ := newUploadTestRouter(t)
	w := uploadFile(t, r, "clip.mov", fakeMOV())
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
}
