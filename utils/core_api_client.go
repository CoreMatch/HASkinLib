package utils

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/lnb/HRPAuth-Backend-Go/config"
	"golang.org/x/oauth2/clientcredentials"
)

type CoreAPIClient struct {
	config *clientcredentials.Config
}

var (
	coreAPIClientInstance *CoreAPIClient
	coreAPIClientOnce     sync.Once
)

func GetCoreAPIClient() *CoreAPIClient {
	coreAPIClientOnce.Do(func() {
		cfg := config.AppConfig.CoreAPI
		if cfg.BaseURL == "" {
			log.Println("CoreAPI base_url is not configured. Service-to-service calls will be disabled.")
			return
		}

		if cfg.ClientID == "" || cfg.ClientSecret == "" {
			log.Println("CoreAPI client_id or client_secret is not configured. Falling back to unauthenticated service calls.")
			coreAPIClientInstance = &CoreAPIClient{}
			return
		}

		coreAPIClientInstance = &CoreAPIClient{
			config: &clientcredentials.Config{
				ClientID:     cfg.ClientID,
				ClientSecret: cfg.ClientSecret,
				TokenURL:     cfg.BaseURL + "/oauth/token",
			},
		}
	})
	return coreAPIClientInstance
}

func (c *CoreAPIClient) GetClient(ctx context.Context) (*http.Client, error) {
	if c == nil || c.config == nil {
		return http.DefaultClient, nil
	}

	return c.config.Client(ctx), nil
}
