package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/database"
	"github.com/lnb/HRPAuth-Backend-Go/models"
)

type ProfileController struct{}

func NewProfileController() *ProfileController {
	return &ProfileController{}
}

type profileTextureItem struct {
	ID          uint   `json:"id"`
	Hash        string `json:"hash"`
	Type        string `json:"type"`
	Model       string `json:"model,omitempty"`
	UID         uint   `json:"uid"`
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

func (pc *ProfileController) GetMyTextures(c *gin.Context) {
	token, err := extractRememberToken(c)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "remember token is required",
		})
		return
	}

	user, err := findUserByRememberToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "invalid remember token",
		})
		return
	}

	textureType := strings.ToLower(strings.TrimSpace(c.Query("type")))
	if textureType == "" {
		textureType = "all"
	}

	var items []profileTextureItem

	if textureType == "all" || textureType == "skin" {
		var skins []models.TextureListSkin
		if err := database.DB.Where("uid = ?", user.UID).Order("id desc").Find(&skins).Error; err == nil {
			for _, skin := range skins {
				items = append(items, profileTextureItem{
					ID:          skin.ID,
					Hash:        skin.Hash,
					Type:        "skin",
					Model:       skin.Model,
					UID:         skin.UID,
					Width:       skin.Width,
					Height:      skin.Height,
					FileName:    skin.FileName,
					PreviewFile: skin.PreviewFile,
					Name:        skin.Name,
					Description: skin.Description,
					Tags:        skin.Tags,
					CreatedAt:   skin.CreatedAt.Format("2006-01-02 15:04:05"),
					UpdatedAt:   skin.UpdatedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}
	}

	if textureType == "all" || textureType == "cape" {
		var capes []models.TextureListCape
		if err := database.DB.Where("uid = ?", user.UID).Order("id desc").Find(&capes).Error; err == nil {
			for _, cape := range capes {
				items = append(items, profileTextureItem{
					ID:          cape.ID,
					Hash:        cape.Hash,
					Type:        "cape",
					UID:         cape.UID,
					Width:       cape.Width,
					Height:      cape.Height,
					FileName:    cape.FileName,
					PreviewFile: cape.PreviewFile,
					Name:        cape.Name,
					Description: cape.Description,
					Tags:        cape.Tags,
					CreatedAt:   cape.CreatedAt.Format("2006-01-02 15:04:05"),
					UpdatedAt:   cape.UpdatedAt.Format("2006-01-02 15:04:05"),
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "profile textures retrieved successfully",
		"data": gin.H{
			"uid":   user.UID,
			"items": items,
		},
	})
}
