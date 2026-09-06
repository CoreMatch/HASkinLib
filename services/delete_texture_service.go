package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/database"
	"github.com/lnb/HRPAuth-Backend-Go/models"
	"gorm.io/gorm"
)

var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type DeleteTextureInput struct {
	Type string
	UID  uint
	Hash string
}

type TextureDeleteService struct{}

func NewTextureDeleteService() *TextureDeleteService {
	return &TextureDeleteService{}
}

// DeleteTexture 删除指定用户指定 type+hash 的纹理记录。
// 同时强制清理磁盘上的原始 PNG 文件和 WebP 预览图,即使其他用户仍引用相同 hash。
func (s *TextureDeleteService) DeleteTexture(input DeleteTextureInput) (*models.TextureRecord, error) {
	textureType := input.Type
	if textureType != "skin" && textureType != "cape" {
		return nil, ErrInvalidTextureType
	}

	hash := input.Hash
	if !hashPattern.MatchString(hash) {
		return nil, ErrInvalidTextureHash
	}

	switch textureType {
	case "skin":
		return deleteSkinTexture(input.UID, hash)
	case "cape":
		return deleteCapeTexture(input.UID, hash)
	default:
		return nil, ErrInvalidTextureType
	}
}

func deleteSkinTexture(uid uint, hash string) (*models.TextureRecord, error) {
	var existing models.TextureListSkin
	err := database.DB.
		Where("uid = ? AND hash = ?", uid, hash).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTextureNotFound
		}
		return nil, fmt.Errorf("failed to query skin texture: %w", err)
	}

	record := toSkinTextureRecord("skin", &existing)

	if deleteErr := database.DB.Delete(&existing).Error; deleteErr != nil {
		return nil, fmt.Errorf("failed to delete skin texture record: %w", deleteErr)
	}

	removeTextureFiles(hash, existing.PreviewFile)

	return record, nil
}

func deleteCapeTexture(uid uint, hash string) (*models.TextureRecord, error) {
	var existing models.TextureListCape
	err := database.DB.
		Where("uid = ? AND hash = ?", uid, hash).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTextureNotFound
		}
		return nil, fmt.Errorf("failed to query cape texture: %w", err)
	}

	record := toCapeTextureRecord("cape", &existing)

	if deleteErr := database.DB.Delete(&existing).Error; deleteErr != nil {
		return nil, fmt.Errorf("failed to delete cape texture record: %w", deleteErr)
	}

	removeTextureFiles(hash, existing.PreviewFile)

	return record, nil
}

// removeTextureFiles 强制清理磁盘上的纹理原文件和预览图。
// 注意:这里采用强制删除策略,即使其他用户仍引用该 hash 也会删除文件。
func removeTextureFiles(hash, previewFileName string) {
	if storageDir := config.AppConfig.Textures.StorageDir; storageDir != "" {
		path := filepath.Join(storageDir, hash)
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("warning: failed to remove texture file %s: %v\n", path, err)
		}
	}

	if previewFileName != "" {
		if previewDir := config.AppConfig.Textures.PreviewStorageDir; previewDir != "" {
			path := filepath.Join(previewDir, previewFileName)
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				fmt.Printf("warning: failed to remove preview file %s: %v\n", path, err)
			}
		}
	}
}