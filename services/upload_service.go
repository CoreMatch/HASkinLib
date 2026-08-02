package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/database"
	"github.com/lnb/HRPAuth-Backend-Go/models"
	"gorm.io/gorm"
)

var (
	ErrTextureFileRequired = errors.New("texture file is required")
	ErrTextureMustBePNG    = errors.New("texture file must be a valid PNG image")
	ErrInvalidTextureType  = errors.New("texture type must be skin or cape")
	ErrInvalidTextureModel = errors.New("texture model must be default or slim")
	ErrTextureNameRequired = errors.New("texture name is required")
)

type UploadTextureInput struct {
	Type             string
	UID              uint
	Model            string
	Name             string
	Description      string
	Tags             string
	OriginalFileName string
	File             io.Reader
}

type TextureUploadService struct{}

func NewTextureUploadService() *TextureUploadService {
	return &TextureUploadService{}
}

func (s *TextureUploadService) UploadTexture(input UploadTextureInput) (*models.TextureList, bool, error) {
	textureType, model, err := normalizeTextureParams(input.Type, input.Model)
	if err != nil {
		return nil, false, err
	}

	if strings.TrimSpace(input.Name) == "" {
		return nil, false, ErrTextureNameRequired
	}

	var fileData []byte
	var width, height int
	fileData, width, height, err = readTextureFile(input.File)
	if err != nil {
		return nil, false, err
	}

	hash := calculateHash(fileData)
	if saveErr := saveTextureFile(hash, fileData); saveErr != nil {
		return nil, false, saveErr
	}

	normalizedTags := normalizeTags(input.Tags)
	fileName := strings.TrimSpace(filepath.Base(input.OriginalFileName))
	if fileName == "" {
		fileName = hash + ".png"
	}

	var existing models.TextureList
	err = database.DB.
		Where("uid = ? AND hash = ? AND type = ?", input.UID, hash, textureType).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("failed to query texture record: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		texture := &models.TextureList{
			Hash:        hash,
			Type:        textureType,
			UID:         input.UID,
			Model:       model,
			Width:       width,
			Height:      height,
			FileName:    fileName,
			Name:        strings.TrimSpace(input.Name),
			Description: strings.TrimSpace(input.Description),
			Tags:        normalizedTags,
		}

		if createErr := database.DB.Create(texture).Error; createErr != nil {
			return nil, false, fmt.Errorf("failed to create texture record: %w", createErr)
		}

		return texture, true, nil
	}

	existing.Model = model
	existing.Width = width
	existing.Height = height
	existing.FileName = fileName
	existing.Name = strings.TrimSpace(input.Name)
	existing.Description = strings.TrimSpace(input.Description)
	existing.Tags = normalizedTags

	if updateErr := database.DB.Save(&existing).Error; updateErr != nil {
		return nil, false, fmt.Errorf("failed to update texture record: %w", updateErr)
	}

	return &existing, false, nil
}

func normalizeTextureParams(textureType, model string) (string, string, error) {
	normalizedType := strings.ToLower(strings.TrimSpace(textureType))
	if normalizedType != "skin" && normalizedType != "cape" {
		return "", "", ErrInvalidTextureType
	}

	normalizedModel := strings.ToLower(strings.TrimSpace(model))
	if normalizedType == "cape" {
		return normalizedType, "default", nil
	}

	if normalizedModel == "" {
		normalizedModel = "default"
	}
	if normalizedModel != "default" && normalizedModel != "slim" {
		return "", "", ErrInvalidTextureModel
	}

	return normalizedType, normalizedModel, nil
}

func readTextureFile(file io.Reader) ([]byte, int, int, error) {
	if file == nil {
		return nil, 0, 0, ErrTextureFileRequired
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read texture file: %w", err)
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, ErrTextureMustBePNG
	}

	return data, cfg.Width, cfg.Height, nil
}

func calculateHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func saveTextureFile(hash string, data []byte) error {
	storageDir := config.AppConfig.Textures.StorageDir
	if storageDir == "" {
		return errors.New("texture storage directory is not configured")
	}

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create texture storage directory: %w", err)
	}

	path := filepath.Join(storageDir, hash)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to inspect texture file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write texture file: %w", err)
	}

	return nil
}

func normalizeTags(tags string) string {
	if strings.TrimSpace(tags) == "" {
		return ""
	}

	replacer := strings.NewReplacer("，", ",", "\n", ",", "\r", ",", "\t", ",", ";", ",", "|", ",")
	normalized := replacer.Replace(tags)

	seen := make(map[string]struct{})
	result := make([]string, 0)
	for part := range strings.SplitSeq(normalized, ",") {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}

	return strings.Join(result, ",")
}
