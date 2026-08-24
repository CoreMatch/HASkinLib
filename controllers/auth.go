package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/database"
	"github.com/lnb/HRPAuth-Backend-Go/models"
	"gorm.io/gorm"
)

// StandardErrorResponse conforms to the HRPAuth error envelope.
type StandardErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Error   string `json:"error"`
	Meta    struct {
		RequestID string `json:"request_id"`
	} `json:"meta"`
}

// AbortWithStandardError sends a standardized error response and aborts the request.
func AbortWithStandardError(c *gin.Context, status int, code string, message string) {
	resp := StandardErrorResponse{
		Success: false,
		Message: message,
		Code:    code,
		Error:   code, // HRPAuth uses code as the error field for compatibility
	}
	// RequestID can be empty if not provided by middleware
	resp.Meta.RequestID = c.GetString("request_id")
	
	c.AbortWithStatusJSON(status, resp)
}

// AuthenticateUser validates the Bearer token and returns the user.
func AuthenticateUser(c *gin.Context) (*models.User, error) {
	tokenStr, err := extractBearerToken(c)
	if err != nil {
		AbortWithStandardError(c, http.StatusUnauthorized, "oauth_login_required", "Bearer token is required")
		return nil, err
	}

	var accessToken models.OAuth2AccessToken
	if err := database.DB.Where("token = ? AND expires_at > ?", tokenStr, time.Now()).First(&accessToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			AbortWithStandardError(c, http.StatusUnauthorized, "oauth_invalid_grant", "Invalid or expired access token")
			return nil, err
		}
		AbortWithStandardError(c, http.StatusInternalServerError, "internal_error", "Failed to validate token")
		return nil, err
	}

	var user models.User
	if err := database.DB.Where("uid = ?", accessToken.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			AbortWithStandardError(c, http.StatusUnauthorized, "user_not_found", "User associated with token not found")
			return nil, err
		}
		AbortWithStandardError(c, http.StatusInternalServerError, "internal_error", "Failed to fetch user")
		return nil, err
	}

	return &user, nil
}

func extractBearerToken(c *gin.Context) (string, error) {
	authorization := strings.TrimSpace(c.GetHeader("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		token := strings.TrimSpace(authorization[len("Bearer "):])
		if token == "" {
			return "", errors.New("empty bearer token")
		}
		return token, nil
	}

	return "", errors.New("missing bearer token")
}
