package services

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
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
	var pngData bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&pngData, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}

	data, width, height, err := readTextureFile(bytes.NewReader(pngData.Bytes()))
	if err != nil {
		t.Fatalf("expected valid png, got error: %v", err)
	}
	if len(data) == 0 || width != 64 || height != 32 {
		t.Fatalf("unexpected texture metadata: len=%d width=%d height=%d", len(data), width, height)
	}

	if _, _, _, err := readTextureFile(strings.NewReader("not-a-png")); err != ErrTextureMustBePNG {
		t.Fatalf("expected ErrTextureMustBePNG, got: %v", err)
	}
}
