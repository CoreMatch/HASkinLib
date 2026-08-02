package controllers

import "testing"

func TestHashRegex(t *testing.T) {
	validHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if !hashRegex.MatchString(validHash) {
		t.Errorf("expected hash %s to be valid", validHash)
	}

	invalidHashes := []string{
		"short",
		"too_long_e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"nothex_e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855_extra",
	}

	for _, h := range invalidHashes {
		if hashRegex.MatchString(h) {
			t.Errorf("expected hash %s to be invalid", h)
		}
	}
}
