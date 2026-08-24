package controllers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/config"
)

var errInvalidPreviewFileName = errors.New("preview_file must be a valid webp file name")

type PreviewFileController struct{}

func NewPreviewFileController() *PreviewFileController {
	return &PreviewFileController{}
}

func (pc *PreviewFileController) Get(c *gin.Context) {
	fileName, err := sanitizePreviewFileName(c.Param("preview_file"))
	if err != nil {
		respondError(c, http.StatusBadRequest, CodeInvalidRequest, err.Error())
		return
	}

	storageDir := config.AppConfig.Textures.PreviewStorageDir
	if strings.TrimSpace(storageDir) == "" {
		respondError(c, http.StatusInternalServerError, CodeStorageNotConfigured, "texture preview storage directory is not configured")
		return
	}

	path := filepath.Join(storageDir, fileName)
	info, statErr := os.Stat(path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			respondError(c, http.StatusNotFound, CodeTexturePreviewNotFound, "preview file not found")
			return
		}

		respondError(c, http.StatusInternalServerError, CodeInternalError, "failed to read preview file")
		return
	}
	if info.IsDir() {
		respondError(c, http.StatusNotFound, CodeTexturePreviewNotFound, "preview file not found")
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Content-Type", "image/webp")
	c.File(path)
}

func sanitizePreviewFileName(fileName string) (string, error) {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return "", errInvalidPreviewFileName
	}
	if filepath.Base(trimmed) != trimmed {
		return "", errInvalidPreviewFileName
	}
	if strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") {
		return "", errInvalidPreviewFileName
	}
	if !strings.HasSuffix(strings.ToLower(trimmed), ".webp") {
		return "", errInvalidPreviewFileName
	}
	return trimmed, nil
}
