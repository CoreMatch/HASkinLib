package controllers

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/config"
)

var (
	errInvalidTextureHash = errors.New("texture hash must be a valid 64-character hex string")
	hashRegex             = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
)

type PullTextureController struct{}

func NewPullTextureController() *PullTextureController {
	return &PullTextureController{}
}

func (pc *PullTextureController) Pull(c *gin.Context) {
	hash := strings.TrimSpace(c.Param("hash"))

	if !hashRegex.MatchString(hash) {
		respondError(c, http.StatusBadRequest, CodeInvalidRequest, errInvalidTextureHash.Error())
		return
	}

	storageDir := config.AppConfig.Textures.StorageDir
	if strings.TrimSpace(storageDir) == "" {
		respondError(c, http.StatusInternalServerError, CodeStorageNotConfigured, "texture storage directory is not configured")
		return
	}

	path := filepath.Join(storageDir, hash)

	info, statErr := os.Stat(path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			respondError(c, http.StatusNotFound, CodeTextureNotFound, "texture file not found")
			return
		}

		respondError(c, http.StatusInternalServerError, CodeInternalError, "failed to read texture file")
		return
	}
	if info.IsDir() {
		respondError(c, http.StatusNotFound, CodeTextureNotFound, "texture file not found")
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Content-Type", "image/png")
	c.File(path)
}
