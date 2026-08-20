package handlers

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// ThumbSuffix marks the tiny blurred placeholder stored next to each photo.
const ThumbSuffix = ".thumb.jpg"

// maxImagePixels bounds decoded photos, rejecting decompression bombs.
const maxImagePixels = 50_000_000

// ThumbDiskPath maps /data/media/<name>.<ext> to /data/media/<name>.thumb.jpg.
func ThumbDiskPath(filePath string) string {
	ext := filepath.Ext(filePath)
	return strings.TrimSuffix(filePath, ext) + ThumbSuffix
}

// ThumbWebPath maps /media/<name>.<ext> to /media/<name>.thumb.jpg.
func ThumbWebPath(webPath string) string {
	ext := filepath.Ext(webPath)
	return strings.TrimSuffix(webPath, ext) + ThumbSuffix
}

// decodeImage reads a photo, checking the header dimensions against
// maxImagePixels before decoding the full image.
func decodeImage(filePath string) (image.Image, error) {
	// #nosec G304 -- filePath is built from a server-generated random
	// filename by the upload handler.
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width*cfg.Height > maxImagePixels {
		return nil, fmt.Errorf("image dimensions %dx%d out of range", cfg.Width, cfg.Height)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return src, nil
}

// thumbMaxSide bounds both sides of the generated placeholder.
const thumbMaxSide = 32

// generateThumb writes a tiny jpeg placeholder next to the photo, scaling by
// the longer side so tall narrow images stay bounded too.
// Best-effort: an encode failure just means the photo has no placeholder,
// the upload itself still succeeds.
func generateThumb(src image.Image, filePath string) error {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w >= h && w > thumbMaxSide {
		h = h * thumbMaxSide / w
		w = thumbMaxSide
	} else if h > thumbMaxSide {
		w = w * thumbMaxSide / h
		h = thumbMaxSide
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	out, err := os.Create(ThumbDiskPath(filePath))
	if err != nil {
		return err
	}
	if err := jpeg.Encode(out, dst, &jpeg.Options{Quality: 50}); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
