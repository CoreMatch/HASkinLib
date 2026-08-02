package controllers

import "testing"

func TestSanitizePreviewFileNameAcceptsGeneratedName(t *testing.T) {
	name, err := sanitizePreviewFileName("8c9b0f_skin.webp")
	if err != nil {
		t.Fatalf("expected generated preview file name to pass, got error: %v", err)
	}
	if name != "8c9b0f_skin.webp" {
		t.Fatalf("unexpected sanitized name: %q", name)
	}
}

func TestSanitizePreviewFileNameRejectsTraversal(t *testing.T) {
	if _, err := sanitizePreviewFileName("../secret.webp"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}

func TestSanitizePreviewFileNameRejectsNonWebP(t *testing.T) {
	if _, err := sanitizePreviewFileName("preview.png"); err == nil {
		t.Fatal("expected non-webp file to be rejected")
	}
}
