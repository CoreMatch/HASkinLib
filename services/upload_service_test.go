package services

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"
)

func TestNormalizeTextureParams(t *testing.T) {
	textureType, model, err := normalizeTextureParams("skin", "slim")
	if err != nil {
		t.Fatalf("expected valid skin params, got error: %v", err)
	}
	if textureType != "skin" || model != "slim" {
		t.Fatalf("unexpected normalized result: %s %s", textureType, model)
	}

	textureType, model, err = normalizeTextureParams("cape", "slim")
	if err != nil {
		t.Fatalf("expected valid cape params, got error: %v", err)
	}
	if textureType != "cape" || model != "default" {
		t.Fatalf("cape model should be normalized to default, got: %s %s", textureType, model)
	}

	if _, _, err := normalizeTextureParams("banner", "default"); err != ErrInvalidTextureType {
		t.Fatalf("expected ErrInvalidTextureType, got: %v", err)
	}

	if _, _, err := normalizeTextureParams("skin", "wide"); err != ErrInvalidTextureModel {
		t.Fatalf("expected ErrInvalidTextureModel, got: %v", err)
	}
}

func TestNormalizeTags(t *testing.T) {
	normalized := normalizeTags(" pvp, 高清 ，pvp\n展示 ; 收藏 ")
	if normalized != "pvp,高清,展示,收藏" {
		t.Fatalf("unexpected normalized tags: %q", normalized)
	}
}

func TestReadTextureFile(t *testing.T) {
	data, width, height, err := readTextureFile(newPNGReader(t, 64, 32), "skin")
	if err != nil {
		t.Fatalf("expected valid png, got error: %v", err)
	}
	if len(data) == 0 || width != 64 || height != 32 {
		t.Fatalf("unexpected texture metadata: len=%d width=%d height=%d", len(data), width, height)
	}

	if _, _, _, err := readTextureFile(strings.NewReader("not-a-png"), "skin"); err != ErrTextureMustBePNG {
		t.Fatalf("expected ErrTextureMustBePNG, got: %v", err)
	}
}

func TestReadTextureFileRejectsInvalidSizes(t *testing.T) {
	if _, _, _, err := readTextureFile(newPNGReader(t, 128, 128), "skin"); err != ErrInvalidSkinSize {
		t.Fatalf("expected ErrInvalidSkinSize, got: %v", err)
	}

	if _, _, _, err := readTextureFile(newPNGReader(t, 32, 32), "cape"); err != ErrInvalidCapeSize {
		t.Fatalf("expected ErrInvalidCapeSize, got: %v", err)
	}
}

func TestReadTextureFileNormalizesLegacyCape(t *testing.T) {
	data, width, height, err := readTextureFile(newPNGReader(t, 22, 17), "cape")
	if err != nil {
		t.Fatalf("expected valid legacy cape, got error: %v", err)
	}
	if width != 64 || height != 32 {
		t.Fatalf("expected normalized cape size 64x32, got %dx%d", width, height)
	}

	normalized, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode normalized png: %v", err)
	}
	if normalized.Width != 64 || normalized.Height != 32 {
		t.Fatalf("expected normalized png canvas 64x32, got %dx%d", normalized.Width, normalized.Height)
	}
}

func TestUploadRateLimiterAllow(t *testing.T) {
	limiter := NewUploadRateLimiter(2, time.Minute)
	now := time.Now()

	if !limiter.Allow("token:test", now) {
		t.Fatal("expected first attempt to pass")
	}
	if !limiter.Allow("token:test", now.Add(10*time.Second)) {
		t.Fatal("expected second attempt to pass")
	}
	if limiter.Allow("token:test", now.Add(20*time.Second)) {
		t.Fatal("expected third attempt within window to be rejected")
	}
	if !limiter.Allow("token:test", now.Add(61*time.Second)) {
		t.Fatal("expected request after window to pass")
	}
}

func newPNGReader(t *testing.T, width, height int) *bytes.Reader {
	t.Helper()

	var pngData bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&pngData, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}

	return bytes.NewReader(pngData.Bytes())
}
