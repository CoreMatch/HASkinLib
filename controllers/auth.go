package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/clients"
	"github.com/lnb/HRPAuth-Backend-Go/models"
)

// AuthenticateUser validates the Bearer token and returns the user.
func AuthenticateUser(c *gin.Context) (*models.AuthenticatedUser, error) {
	tokenStr, err := extractBearerToken(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "oauth_login_required", "Bearer token is required")
		return nil, err
	}

	coreClient := clients.NewCoreClient()
	if coreClient.BaseURL == "" {
		respondError(c, http.StatusInternalServerError, CodeInternalError, "Database error during user retrieval")
		return nil, errors.New("missing core api base url")
	}

	user, err := coreClient.VerifyToken(tokenStr)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "oauth_invalid_grant", "Invalid, expired, or revoked access token")
		return nil, err
	}

	return user, nil
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
