package controllers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseListPreviewOptionsDefaults(t *testing.T) {
	c := newListPreviewTestContext("/texture/listpreview")

	opts, err := parseListPreviewOptions(c)
	if err != nil {
		t.Fatalf("expected defaults to parse, got error: %v", err)
	}

	if opts.Filter != "all" {
		t.Fatalf("expected default filter all, got %q", opts.Filter)
	}
	if opts.Order != "desc" {
		t.Fatalf("expected default order desc, got %q", opts.Order)
	}
	if opts.Page != 1 {
		t.Fatalf("expected default page 1, got %d", opts.Page)
	}
	if opts.Limit != defaultListPreviewPageSize {
		t.Fatalf("expected default limit %d, got %d", defaultListPreviewPageSize, opts.Limit)
	}
	if opts.Offset != 0 {
		t.Fatalf("expected default offset 0, got %d", opts.Offset)
	}
}

func TestParseListPreviewOptionsCustomValues(t *testing.T) {
	c := newListPreviewTestContext("/texture/listpreview?type=slim&order=asc&page=3&tag=%E5%B1%95%E7%A4%BA")

	opts, err := parseListPreviewOptions(c)
	if err != nil {
		t.Fatalf("expected custom values to parse, got error: %v", err)
	}

	if opts.Filter != "slim" {
		t.Fatalf("expected filter slim, got %q", opts.Filter)
	}
	if opts.Order != "asc" {
		t.Fatalf("expected order asc, got %q", opts.Order)
	}
	if opts.Page != 3 {
		t.Fatalf("expected page 3, got %d", opts.Page)
	}
	if opts.Offset != 32 {
		t.Fatalf("expected offset 32, got %d", opts.Offset)
	}
	if opts.Tag != "展示" {
		t.Fatalf("expected tag 展示, got %q", opts.Tag)
	}
}

func TestParseListPreviewOptionsRejectsInvalidType(t *testing.T) {
	c := newListPreviewTestContext("/texture/listpreview?type=skin")

	if _, err := parseListPreviewOptions(c); err == nil {
		t.Fatal("expected invalid type to be rejected")
	}
}

func TestParseListPreviewOptionsRejectsInvalidOrder(t *testing.T) {
	c := newListPreviewTestContext("/texture/listpreview?order=newest")

	if _, err := parseListPreviewOptions(c); err == nil {
		t.Fatal("expected invalid order to be rejected")
	}
}

func TestParseListPreviewOptionsRejectsInvalidPage(t *testing.T) {
	c := newListPreviewTestContext("/texture/listpreview?page=0")

	if _, err := parseListPreviewOptions(c); err == nil {
		t.Fatal("expected invalid page to be rejected")
	}
}

func TestNormalizeListPreviewTag(t *testing.T) {
	tag := normalizeListPreviewTag("  收藏，展示 ")
	if tag != "收藏" {
		t.Fatalf("expected normalized tag 收藏, got %q", tag)
	}
}

func newListPreviewTestContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", target, nil)
	return c
}
