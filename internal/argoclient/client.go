package argoclient

import (
	"fmt"
	"os"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
)

// Config defines the configuration for creating an Argo CD client
type Config struct {
	Server    string
	AuthToken string
	Insecure  bool
}

// NewClient creates a new Argo CD API client with the provided configuration
func NewClient(cfg Config) (apiclient.Client, error) {
	// Get server address from config or environment
	server := cfg.Server
	if server == "" {
		server = os.Getenv("ARGOCD_BASE_URL")
		if server == "" {
			return nil, fmt.Errorf("argo CD server address not provided. Set ARGOCD_BASE_URL env var or pass server in config")
		}
	}

	// Strip URL scheme if present (https:// or http://)
	// The Argo CD gRPC client expects just the hostname:port
	server = strings.TrimPrefix(server, "https://")
	server = strings.TrimPrefix(server, "http://")

	// Get auth token from config or environment
	authToken := cfg.AuthToken
	if authToken == "" {
		authToken = os.Getenv("ARGOCD_API_TOKEN")
		if authToken == "" {
			return nil, fmt.Errorf("argo CD auth token not provided. Set ARGOCD_API_TOKEN env var or pass auth_token in config")
		}
	}

	// Create Argo CD client
	clientOpts := apiclient.ClientOptions{
		ServerAddr: server,
		AuthToken:  authToken,
		Insecure:   cfg.Insecure,
	}

	apiClient, err := apiclient.NewClient(&clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Argo CD client: %w", err)
	}

	return apiClient, nil
}
