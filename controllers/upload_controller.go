package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/models"
	"github.com/lnb/HRPAuth-Backend-Go/services"
)

type UploadController struct {
	uploadService *services.TextureUploadService
	rateLimiter   *services.UploadRateLimiter
}

const (
	defaultMaxTextureUploadBytes        int64 = 2 << 20
	defaultMaxTextureRequestBytes       int64 = defaultMaxTextureUploadBytes + (256 << 10)
	defaultUploadRateLimitPerMinute           = 5
	defaultUploadRateLimitWindowSeconds       = 60
)

func NewUploadController() *UploadController {
	textureCfg := config.AppConfig.Textures
	return &UploadController{
		uploadService: services.NewTextureUploadService(),
		rateLimiter: services.NewUploadRateLimiter(
			getUploadRateLimitPerMinute(textureCfg),
			time.Duration(getUploadRateLimitWindowSeconds(textureCfg))*time.Second,
		),
	}
}

type uploadTextureResponse struct {
	ID          uint   `json:"id"`
	Hash        string `json:"hash"`
	Type        string `json:"type"`
	UID         uint   `json:"uid"`
	Model       string `json:"model,omitempty"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FileName    string `json:"file_name"`
	PreviewFile string `json:"preview_file"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (uc *UploadController) UploadTexture(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, getMaxTextureRequestBytes())

	user, err := AuthenticateUser(c)
	if err != nil {
		// AuthenticateUser already handles response and aborts
		return
	}

	accessToken, _ := extractBearerToken(c) // already validated by AuthenticateUser
	now := time.Now()
	if !uc.rateLimiter.Allow("token:"+accessToken, now) || !uc.rateLimiter.Allow("ip:"+c.ClientIP(), now) {
		AbortWithStandardError(c, http.StatusTooManyRequests, "upload_rate_limited", "Upload rate limit exceeded")
		return
	}

	uidValue := strings.TrimSpace(c.PostForm("uid"))
	uid64, err := strconv.ParseUint(uidValue, 10, 64)
	if err != nil || uid64 == 0 {
		AbortWithStandardError(c, http.StatusBadRequest, "invalid_request", "uid must be a positive integer")
		return
	}

	if uint(uid64) != user.UID {
		AbortWithStandardError(c, http.StatusForbidden, "oauth_access_denied", "Token does not match the requested uid")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fileHeader, err = c.FormFile("texture")
		if err != nil {
			if isBodyTooLargeError(err) {
				AbortWithStandardError(c, http.StatusRequestEntityTooLarge, "upload_request_too_large", "Upload request is too large")
				return
			}

			AbortWithStandardError(c, http.StatusBadRequest, "texture_file_required", "Texture file is required")
			return
		}
	}

	if fileHeader.Size > getMaxTextureUploadBytes() {
		AbortWithStandardError(c, http.StatusRequestEntityTooLarge, "upload_request_too_large", fmt.Sprintf("Texture file must be %d bytes or smaller", getMaxTextureUploadBytes()))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		AbortWithStandardError(c, http.StatusBadRequest, "invalid_texture_file", "Failed to open uploaded file")
		return
	}
	defer file.Close()

	texture, created, err := uc.uploadService.UploadTexture(services.UploadTextureInput{
		Type:             c.PostForm("type"),
		UID:              uint(uid64),
		Model:            c.PostForm("model"),
		Name:             c.PostForm("name"),
		Description:      c.PostForm("description"),
		Tags:             c.PostForm("tags"),
		OriginalFileName: fileHeader.Filename,
		File:             file,
	})
	if err != nil {
		status := http.StatusInternalServerError
		code := "texture_upload_failed"
		switch {
		case errors.Is(err, services.ErrInvalidTextureType):
			status, code = http.StatusBadRequest, "invalid_texture_type"
		case errors.Is(err, services.ErrInvalidTextureModel):
			status, code = http.StatusBadRequest, "invalid_texture_model"
		case errors.Is(err, services.ErrTextureMustBePNG):
			status, code = http.StatusBadRequest, "invalid_texture_file"
		case errors.Is(err, services.ErrInvalidSkinSize), errors.Is(err, services.ErrInvalidCapeSize):
			status, code = http.StatusBadRequest, "invalid_texture_size"
		case errors.Is(err, services.ErrTextureNameRequired):
			status, code = http.StatusBadRequest, "texture_name_required"
		case errors.Is(err, services.ErrTextureFileRequired):
			status, code = http.StatusBadRequest, "texture_file_required"
		}

		AbortWithStandardError(c, status, code, err.Error())
		return
	}

	message := "texture uploaded successfully"
	status := http.StatusCreated
	if !created {
		message = "texture metadata updated successfully"
		status = http.StatusOK
	}

	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    buildUploadTextureResponse(texture),
	})
}

func buildUploadTextureResponse(texture *models.TextureRecord) uploadTextureResponse {
	return uploadTextureResponse{
		ID:          texture.ID,
		Hash:        texture.Hash,
		Type:        texture.Type,
		UID:         texture.UID,
		Model:       texture.Model,
		Width:       texture.Width,
		Height:      texture.Height,
		FileName:    texture.FileName,
		PreviewFile: texture.PreviewFile,
		Name:        texture.Name,
		Description: texture.Description,
		Tags:        texture.Tags,
		CreatedAt:   texture.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   texture.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func isBodyTooLargeError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, http.ErrBodyReadAfterClose) {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "request body too large")
}

func getMaxTextureUploadBytes() int64 {
	if config.AppConfig == nil || config.AppConfig.Textures.MaxUploadBytes <= 0 {
		return defaultMaxTextureUploadBytes
	}
	return config.AppConfig.Textures.MaxUploadBytes
}

func getMaxTextureRequestBytes() int64 {
	if config.AppConfig == nil {
		return defaultMaxTextureRequestBytes
	}

	maxRequestBytes := config.AppConfig.Textures.MaxRequestBytes
	if maxRequestBytes <= 0 {
		return defaultMaxTextureRequestBytes
	}
	maxUploadBytes := getMaxTextureUploadBytes()
	if maxRequestBytes < maxUploadBytes {
		return maxUploadBytes
	}
	return maxRequestBytes
}

func getUploadRateLimitPerMinute(textureCfg config.TextureConfig) int {
	if textureCfg.RateLimitPerMinute <= 0 {
		return defaultUploadRateLimitPerMinute
	}
	return textureCfg.RateLimitPerMinute
}

func getUploadRateLimitWindowSeconds(textureCfg config.TextureConfig) int {
	if textureCfg.RateLimitWindowSeconds <= 0 {
		return defaultUploadRateLimitWindowSeconds
	}
	return textureCfg.RateLimitWindowSeconds
}
