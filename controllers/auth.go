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

// AuthenticateUser validates the Bearer token and returns the user.
func AuthenticateUser(c *gin.Context) (*models.User, error) {
	tokenStr, err := extractBearerToken(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "oauth_login_required", "Bearer token is required")
		return nil, err
	}

	var accessToken models.OAuth2AccessToken
	if err := database.DB.Where("access_token = ? AND expires_at > ? AND revoked_at IS NULL", tokenStr, time.Now()).First(&accessToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusUnauthorized, "oauth_invalid_grant", "Invalid, expired, or revoked access token")
			return nil, err
		}
		respondError(c, http.StatusInternalServerError, CodeInternalError, "Failed to validate token")
		return nil, err
	}

	var user models.User
	var query *gorm.DB
	if accessToken.SubjectType == "user" && accessToken.UserID != nil {
		query = database.DB.Where("uuid = ?", *accessToken.UserID)
	} else if accessToken.SubjectType == "service" && accessToken.TargetUID != nil {
		query = database.DB.Where("uid = ?", *accessToken.TargetUID)
	} else {
		respondError(c, http.StatusUnauthorized, "oauth_invalid_grant", "Token does not have a valid user subject")
		return nil, errors.New("invalid token subject")
	}

	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(c, http.StatusUnauthorized, "user_not_found", "User associated with token not found")
			return nil, err
		}
		respondError(c, http.StatusInternalServerError, CodeInternalError, "Failed to fetch user")
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
