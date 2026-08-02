package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/database"
	"github.com/lnb/HRPAuth-Backend-Go/models"
)

const defaultListPreviewPageSize = 16

type ListPreviewController struct{}

func NewListPreviewController() *ListPreviewController {
	return &ListPreviewController{}
}

type listPreviewOptions struct {
	Filter string
	Order  string
	Tag    string
	Page   int
	Limit  int
	Offset int
}

type listPreviewItem struct {
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

func (lc *ListPreviewController) List(c *gin.Context) {
	opts, err := parseListPreviewOptions(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	items, total, err := lc.queryPreviewItems(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to query texture preview list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "texture preview list retrieved successfully",
		"data": gin.H{
			"items":     items,
			"filter":    opts.Filter,
			"order":     opts.Order,
			"tag":       opts.Tag,
			"page":      opts.Page,
			"page_size": opts.Limit,
			"total":     total,
			"has_more":  int64(opts.Offset+len(items)) < total,
		},
	})
}

func (lc *ListPreviewController) queryPreviewItems(opts listPreviewOptions) ([]listPreviewItem, int64, error) {
	switch opts.Filter {
	case "cape":
		return queryCapePreviewItems(opts)
	case "default", "slim":
		return querySkinPreviewItems(opts, opts.Filter)
	case "all":
		return querySkinPreviewItems(opts, "")
	default:
		return nil, 0, nil
	}
}

func querySkinPreviewItems(opts listPreviewOptions, modelFilter string) ([]listPreviewItem, int64, error) {
	query := database.DB.Model(&models.TextureListSkin{})
	if modelFilter != "" {
		query = query.Where("model = ?", modelFilter)
	}
	if opts.Tag != "" {
		query = query.Where("FIND_IN_SET(?, tags) > 0", opts.Tag)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []models.TextureListSkin
	if err := query.Order("id " + opts.Order).Limit(opts.Limit).Offset(opts.Offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	items := make([]listPreviewItem, 0, len(records))
	for _, record := range records {
		items = append(items, listPreviewItem{
			ID:          record.ID,
			Hash:        record.Hash,
			Type:        "skin",
			Model:       record.Model,
			UID:         record.UID,
			Width:       record.Width,
			Height:      record.Height,
			FileName:    record.FileName,
			PreviewFile: record.PreviewFile,
			Name:        record.Name,
			Description: record.Description,
			Tags:        record.Tags,
			CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return items, total, nil
}

func queryCapePreviewItems(opts listPreviewOptions) ([]listPreviewItem, int64, error) {
	query := database.DB.Model(&models.TextureListCape{})
	if opts.Tag != "" {
		query = query.Where("FIND_IN_SET(?, tags) > 0", opts.Tag)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []models.TextureListCape
	if err := query.Order("id " + opts.Order).Limit(opts.Limit).Offset(opts.Offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	items := make([]listPreviewItem, 0, len(records))
	for _, record := range records {
		items = append(items, listPreviewItem{
			ID:          record.ID,
			Hash:        record.Hash,
			Type:        "cape",
			UID:         record.UID,
			Width:       record.Width,
			Height:      record.Height,
			FileName:    record.FileName,
			PreviewFile: record.PreviewFile,
			Name:        record.Name,
			Description: record.Description,
			Tags:        record.Tags,
			CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return items, total, nil
}

func parseListPreviewOptions(c *gin.Context) (listPreviewOptions, error) {
	filter := strings.ToLower(strings.TrimSpace(c.DefaultQuery("type", "all")))
	switch filter {
	case "all", "default", "slim", "cape":
	default:
		return listPreviewOptions{}, errInvalidListPreviewOption("type must be one of all, default, slim, cape")
	}

	order := strings.ToLower(strings.TrimSpace(c.DefaultQuery("order", "desc")))
	switch order {
	case "desc", "asc":
	default:
		return listPreviewOptions{}, errInvalidListPreviewOption("order must be asc or desc")
	}

	page := 1
	if rawPage := strings.TrimSpace(c.Query("page")); rawPage != "" {
		parsedPage, err := strconv.Atoi(rawPage)
		if err != nil || parsedPage <= 0 {
			return listPreviewOptions{}, errInvalidListPreviewOption("page must be a positive integer")
		}
		page = parsedPage
	}

	tag := normalizeListPreviewTag(c.Query("tag"))

	return listPreviewOptions{
		Filter: filter,
		Order:  order,
		Tag:    tag,
		Page:   page,
		Limit:  defaultListPreviewPageSize,
		Offset: (page - 1) * defaultListPreviewPageSize,
	}, nil
}

func normalizeListPreviewTag(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.ReplaceAll(tag, "，", ",")
	if idx := strings.Index(tag, ","); idx >= 0 {
		tag = tag[:idx]
	}
	return strings.TrimSpace(tag)
}

type errInvalidListPreviewOption string

func (e errInvalidListPreviewOption) Error() string {
	return string(e)
}
