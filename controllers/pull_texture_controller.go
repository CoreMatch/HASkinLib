package controllers

import (
	"bytes"
	"encoding/json"
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

	// #region debug-point A:pull-start
	go func() {
		d := map[string]interface{}{"sessionId": "texture-pull-404-error", "runId": "pre", "hypothesisId": "A", "location": "pull_texture_controller.go:27", "msg": "[DEBUG] Pull start", "data": map[string]interface{}{"hash": hash, "storageDir": config.AppConfig.Textures.StorageDir}}
		b, _ := json.Marshal(d)
		http.Post("http://127.0.0.1:7777/event", "application/json", bytes.NewBuffer(b))
	}()
	// #endregion

	if !hashRegex.MatchString(hash) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": errInvalidTextureHash.Error(),
		})
		return
	}

	storageDir := config.AppConfig.Textures.StorageDir
	if strings.TrimSpace(storageDir) == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "texture storage directory is not configured",
		})
		return
	}

	path := filepath.Join(storageDir, hash)

	// #region debug-point B:stat-check
	go func() {
		entries, _ := os.ReadDir(storageDir)
		fileList := []string{}
		for _, e := range entries {
			fileList = append(fileList, e.Name())
		}
		d := map[string]interface{}{"sessionId": "texture-pull-404-error", "runId": "pre", "hypothesisId": "B", "location": "pull_texture_controller.go:49", "msg": "[DEBUG] Stat check", "data": map[string]interface{}{"path": path, "exists": false, "dirContents": fileList}}
		if _, err := os.Stat(path); err == nil {
			d["data"].(map[string]interface{})["exists"] = true
		}
		b, _ := json.Marshal(d)
		http.Post("http://127.0.0.1:7777/event", "application/json", bytes.NewBuffer(b))
	}()
	// #endregion

	info, statErr := os.Stat(path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			// #region debug-point B:not-found
			go func() {
				d := map[string]interface{}{"sessionId": "texture-pull-404-error", "runId": "pre", "hypothesisId": "B", "location": "pull_texture_controller.go:54", "msg": "[DEBUG] File not found error", "data": map[string]interface{}{"path": path}}
				b, _ := json.Marshal(d)
				http.Post("http://127.0.0.1:7777/event", "application/json", bytes.NewBuffer(b))
			}()
			// #endregion
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "texture file not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to read texture file",
		})
		return
	}
	if info.IsDir() {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "texture file not found",
		})
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("Content-Type", "image/png")
	c.File(path)
}
