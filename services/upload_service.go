package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/HugoSmits86/nativewebp"
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
	ErrInvalidSkinSize     = errors.New("skin texture must be 64x32 or 64x64")
	ErrInvalidCapeSize     = errors.New("cape texture must be 64x32 or 22x17")
)

const previewScale = 8

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
	fileData, width, height, err = readTextureFile(input.File, textureType)
	if err != nil {
		return nil, false, err
	}

	hash := calculateHash(fileData)
	previewData, err := generatePreviewImage(fileData, textureType, model)
	if err != nil {
		return nil, false, fmt.Errorf("failed to generate texture preview: %w", err)
	}

	if saveErr := saveTextureFile(hash, fileData); saveErr != nil {
		return nil, false, saveErr
	}
	previewFileName := buildPreviewFileName(hash, textureType)
	if saveErr := savePreviewFile(previewFileName, previewData); saveErr != nil {
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
			PreviewFile: previewFileName,
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
	existing.PreviewFile = previewFileName
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

func readTextureFile(file io.Reader, textureType string) ([]byte, int, int, error) {
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

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, ErrTextureMustBePNG
	}

	width := cfg.Width
	height := cfg.Height

	switch textureType {
	case "skin":
		if !isValidSkinSize(width, height) {
			return nil, 0, 0, ErrInvalidSkinSize
		}
	case "cape":
		if !isValidCapeSize(width, height) {
			return nil, 0, 0, ErrInvalidCapeSize
		}
		if width == 22 && height == 17 {
			img = normalizeCapeImage(img)
			width = 64
			height = 32
		}
	default:
		return nil, 0, 0, ErrInvalidTextureType
	}

	normalizedData, err := encodePNG(img)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to encode normalized texture: %w", err)
	}

	return normalizedData, width, height, nil
}

func isValidSkinSize(width, height int) bool {
	return (width == 64 && height == 32) || (width == 64 && height == 64)
}

func isValidCapeSize(width, height int) bool {
	return (width == 64 && height == 32) || (width == 22 && height == 17)
}

func normalizeCapeImage(img image.Image) image.Image {
	normalized := image.NewRGBA(image.Rect(0, 0, 64, 32))
	draw.Draw(normalized, normalized.Bounds(), image.Transparent, image.Point{}, draw.Src)
	draw.Draw(normalized, img.Bounds(), img, image.Point{}, draw.Src)
	return normalized
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeWebP(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, img, &nativewebp.Options{
		CompressionLevel: nativewebp.BestCompression,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func generatePreviewImage(fileData []byte, textureType, model string) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(fileData))
	if err != nil {
		return nil, ErrTextureMustBePNG
	}

	var preview image.Image
	switch textureType {
	case "skin":
		preview = renderSkinPreview(img, model == "slim")
	case "cape":
		preview = renderCapePreview(img)
	default:
		return nil, ErrInvalidTextureType
	}

	return encodeWebP(preview)
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

func savePreviewFile(fileName string, data []byte) error {
	storageDir := config.AppConfig.Textures.PreviewStorageDir
	if storageDir == "" {
		return errors.New("texture preview storage directory is not configured")
	}

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create texture preview directory: %w", err)
	}

	path := filepath.Join(storageDir, fileName)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write texture preview file: %w", err)
	}

	return nil
}

func buildPreviewFileName(hash, textureType string) string {
	return hash + "_" + textureType + ".webp"
}

func renderSkinPreview(src image.Image, slim bool) image.Image {
	armWidth := 4
	if slim {
		armWidth = 3
	}

	canvasWidth := 8 + (armWidth * 2)
	base := image.NewNRGBA(image.Rect(0, 0, canvasWidth, 32))
	draw.Draw(base, base.Bounds(), image.Transparent, image.Point{}, draw.Src)

	torsoX := armWidth
	headX := torsoX
	leftArmX := 0
	rightArmX := torsoX + 8
	leftLegX := torsoX
	rightLegX := torsoX + 4

	drawPart(base, src, image.Rect(8, 8, 16, 16), image.Pt(headX, 0))
	overlayPart(base, src, image.Rect(40, 8, 48, 16), image.Pt(headX, 0))

	drawPart(base, src, image.Rect(20, 20, 28, 32), image.Pt(torsoX, 8))
	overlayPart(base, src, image.Rect(20, 36, 28, 48), image.Pt(torsoX, 8))

	armFront := image.Rect(44, 20, 44+armWidth, 32)
	armOverlay := image.Rect(44, 36, 44+armWidth, 48)
	drawPart(base, src, armFront, image.Pt(leftArmX, 8))
	drawPart(base, src, armFront, image.Pt(rightArmX, 8))
	overlayPart(base, src, armOverlay, image.Pt(leftArmX, 8))
	overlayPart(base, src, armOverlay, image.Pt(rightArmX, 8))

	legFront := image.Rect(4, 20, 8, 32)
	legOverlay := image.Rect(4, 36, 8, 48)
	drawPart(base, src, legFront, image.Pt(leftLegX, 20))
	drawPart(base, src, legFront, image.Pt(rightLegX, 20))
	overlayPart(base, src, legOverlay, image.Pt(leftLegX, 20))
	overlayPart(base, src, legOverlay, image.Pt(rightLegX, 20))

	return scaleNearest(base, previewScale)
}

func renderCapePreview(src image.Image) image.Image {
	base := image.NewNRGBA(image.Rect(0, 0, 10, 16))
	draw.Draw(base, base.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// The 10x16 outer cape face is the most recognizable flat preview area.
	drawPart(base, src, image.Rect(1, 1, 11, 17), image.Point{})
	return scaleNearest(base, previewScale)
}

func drawPart(dst draw.Image, src image.Image, srcRect image.Rectangle, dstMin image.Point) {
	draw.Draw(dst, image.Rectangle{Min: dstMin, Max: dstMin.Add(srcRect.Size())}, src, srcRect.Min, draw.Src)
}

func overlayPart(dst draw.Image, src image.Image, srcRect image.Rectangle, dstMin image.Point) {
	if !rectFits(src.Bounds(), srcRect) {
		return
	}

	draw.Draw(dst, image.Rectangle{Min: dstMin, Max: dstMin.Add(srcRect.Size())}, src, srcRect.Min, draw.Over)
}

func rectFits(bounds, rect image.Rectangle) bool {
	return rect.Min.X >= bounds.Min.X &&
		rect.Min.Y >= bounds.Min.Y &&
		rect.Max.X <= bounds.Max.X &&
		rect.Max.Y <= bounds.Max.Y
}

func scaleNearest(src image.Image, factor int) image.Image {
	if factor <= 1 {
		return src
	}

	srcBounds := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, srcBounds.Dx()*factor, srcBounds.Dy()*factor))
	for y := 0; y < srcBounds.Dy(); y++ {
		for x := 0; x < srcBounds.Dx(); x++ {
			c := color.NRGBAModel.Convert(src.At(srcBounds.Min.X+x, srcBounds.Min.Y+y)).(color.NRGBA)
			fillScaledPixel(dst, x*factor, y*factor, factor, c)
		}
	}
	return dst
}

func fillScaledPixel(dst *image.NRGBA, startX, startY, size int, c color.NRGBA) {
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dst.SetNRGBA(startX+x, startY+y, c)
		}
	}
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
