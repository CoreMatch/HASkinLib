package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDContextKey = "request_id"

const (
	CodeInvalidRequest         = "invalid_request"
	CodeUploadRequestTooLarge  = "upload_request_too_large"
	CodeUploadRateLimited      = "upload_rate_limited"
	CodeTextureTypeInvalid     = "invalid_texture_type"
	CodeTextureModelInvalid    = "invalid_texture_model"
	CodeTextureFileRequired    = "texture_file_required"
	CodeTextureFileInvalid     = "invalid_texture_file"
	CodeTextureNameRequired    = "texture_name_required"
	CodeTextureSizeInvalid     = "invalid_texture_size"
	CodeTextureUploadFailed    = "texture_upload_failed"
	CodeTexturePreviewNotFound = "preview_not_found"
	CodeTextureNotFound        = "texture_not_found"
	CodeProfileQueryFailed     = "profile_query_failed"
	CodeStorageNotConfigured   = "storage_not_configured"
	CodeInternalError          = "internal_error"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-Id"))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(requestIDContextKey, requestID)
		c.Header("X-Request-Id", requestID)
		c.Next()
	}
}

func respondOK(c *gin.Context, message string, data any) {
	respondJSON(c, http.StatusOK, true, "", message, data)
}

func respondCreated(c *gin.Context, message string, data any) {
	respondJSON(c, http.StatusCreated, true, "", message, data)
}

func respondError(c *gin.Context, status int, code, message string) {
	respondJSON(c, status, false, code, message, nil)
}

func respondJSON(c *gin.Context, status int, success bool, code, message string, data any) {
	payload := gin.H{
		"success": success,
		"meta": gin.H{
			"request_id": requestIDFrom(c),
		},
	}

	if message != "" {
		payload["message"] = message
	}
	if code != "" {
		payload["code"] = code
		payload["error"] = code
	}
	if data != nil {
		payload["data"] = data
		if success {
			switch typed := data.(type) {
			case gin.H:
				for key, value := range typed {
					payload[key] = value
				}
			case map[string]any:
				for key, value := range typed {
					payload[key] = value
				}
			}
		}
	}

	c.JSON(status, payload)
}

func requestIDFrom(c *gin.Context) string {
	if value, exists := c.Get(requestIDContextKey); exists {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return newRequestID()
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
