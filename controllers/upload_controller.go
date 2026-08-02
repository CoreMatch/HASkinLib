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
	"github.com/lnb/HRPAuth-Backend-Go/database"
	"github.com/lnb/HRPAuth-Backend-Go/models"
	"github.com/lnb/HRPAuth-Backend-Go/services"
	"gorm.io/gorm"
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
	Model       string `json:"model"`
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

	rememberToken, err := extractRememberToken(c)
	if err != nil {
		if isBodyTooLargeError(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"message": "upload request is too large",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "failed to parse upload form",
		})
		return
	}
	if rememberToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "remember token is required",
		})
		return
	}

	now := time.Now()
	if !uc.rateLimiter.Allow("token:"+rememberToken, now) || !uc.rateLimiter.Allow("ip:"+c.ClientIP(), now) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "upload rate limit exceeded, please try again later",
		})
		return
	}

	user, err := findUserByRememberToken(rememberToken)
	if err != nil {
		status := http.StatusInternalServerError
		message := "failed to authenticate user"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusUnauthorized
			message = "invalid remember token"
		}

		c.JSON(status, gin.H{
			"success": false,
			"message": message,
		})
		return
	}

	uidValue := strings.TrimSpace(c.PostForm("uid"))
	uid64, err := strconv.ParseUint(uidValue, 10, 64)
	if err != nil || uid64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "uid must be a positive integer",
		})
		return
	}

	if uint(uid64) != user.UID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "remember token does not match the requested uid",
		})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		fileHeader, err = c.FormFile("texture")
		if err != nil {
			if isBodyTooLargeError(err) {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{
					"success": false,
					"message": "upload request is too large",
				})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "file is required",
			})
			return
		}
	}

	if fileHeader.Size > getMaxTextureUploadBytes() {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"success": false,
			"message": fmt.Sprintf("texture file must be %d bytes or smaller", getMaxTextureUploadBytes()),
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		if isBodyTooLargeError(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"success": false,
				"message": "upload request is too large",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "failed to open uploaded file",
		})
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
		switch {
		case errors.Is(err, services.ErrInvalidTextureType),
			errors.Is(err, services.ErrInvalidTextureModel),
			errors.Is(err, services.ErrTextureMustBePNG),
			errors.Is(err, services.ErrInvalidSkinSize),
			errors.Is(err, services.ErrInvalidCapeSize),
			errors.Is(err, services.ErrTextureNameRequired),
			errors.Is(err, services.ErrTextureFileRequired):
			status = http.StatusBadRequest
		}

		c.JSON(status, gin.H{
			"success": false,
			"message": err.Error(),
		})
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

func extractRememberToken(c *gin.Context) (string, error) {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		return strings.TrimSpace(authorization[len("Bearer "):]), nil
	}

	if strings.HasPrefix(c.ContentType(), "multipart/") {
		if err := c.Request.ParseMultipartForm(getMaxTextureRequestBytes()); err != nil {
			return "", err
		}
	}

	for _, key := range []string{"remember_token", "rt", "token"} {
		if value := strings.TrimSpace(c.PostForm(key)); value != "" {
			return value, nil
		}
	}

	return "", nil
}

func findUserByRememberToken(rememberToken string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("remember_token = ?", rememberToken).First(&user).Error; err != nil {
		return nil, fmt.Errorf("query user by remember token: %w", err)
	}
	return &user, nil
}

func buildUploadTextureResponse(texture *models.TextureList) uploadTextureResponse {
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
