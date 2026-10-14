package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/models"
	"github.com/lnb/HRPAuth-Backend-Go/utils"
)

type CoreClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewCoreClient() *CoreClient {
	apiClient := utils.GetCoreAPIClient()
	httpClient, _ := apiClient.GetClient(context.Background())

	return &CoreClient{
		BaseURL:    strings.TrimRight(config.AppConfig.CoreAPI.BaseURL, "/"),
		HTTPClient: httpClient,
	}
}

func (c *CoreClient) doRequest(method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("core api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *CoreClient) VerifyToken(token string) (*models.AuthenticatedUser, error) {
	var resp struct {
		Success bool                    `json:"success"`
		Data    models.CoreUserIdentity `json:"data"`
	}

	if err := c.doRequest(http.MethodPost, "/internal/verify-token", map[string]string{
		"token": token,
	}, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("token verification failed")
	}

	return &models.AuthenticatedUser{
		UID:   resp.Data.UID,
		UUID:  resp.Data.UUID,
		Email: resp.Data.Email,
	}, nil
}
