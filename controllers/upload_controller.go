package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/database"
	"github.com/lnb/HRPAuth-Backend-Go/models"
	"github.com/lnb/HRPAuth-Backend-Go/services"
	"gorm.io/gorm"
)

type UploadController struct {
	uploadService *services.TextureUploadService
}

func NewUploadController() *UploadController {
	return &UploadController{
		uploadService: services.NewTextureUploadService(),
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
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (uc *UploadController) UploadTexture(c *gin.Context) {
	rememberToken := extractRememberToken(c)
	if rememberToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "remember token is required",
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
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "file is required",
			})
			return
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
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

func extractRememberToken(c *gin.Context) string {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		return strings.TrimSpace(authorization[len("Bearer "):])
	}

	for _, key := range []string{"remember_token", "rt", "token"} {
		if value := strings.TrimSpace(c.PostForm(key)); value != "" {
			return value
		}
	}

	return ""
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
		Name:        texture.Name,
		Description: texture.Description,
		Tags:        texture.Tags,
		CreatedAt:   texture.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   texture.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
