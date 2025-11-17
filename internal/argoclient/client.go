package argoclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/sethvargo/go-envconfig"
)

// Config defines the configuration for creating an Argo CD client
type Config struct {
	Server    string `env:"ARGOCD_BASE_URL,required"`
	AuthToken string `env:"ARGOCD_API_TOKEN,required"`
	Insecure  bool   `env:"ARGOCD_INSECURE,default=false"`
}

// NewConfigFromEnv loads the Argo CD configuration from environment variables
func NewConfigFromEnv(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}
	return &cfg, nil
}

// ClientWithServer wraps an Argo CD client with its server URL
type ClientWithServer struct {
	Client apiclient.Client
	Server string
}

// NewClient creates a new Argo CD API client with the provided configuration
// Returns the client and the normalized server URL it's connected to
func NewClient(cfg Config) (*ClientWithServer, error) {
	// Strip URL scheme if present (https:// or http://)
	// The Argo CD gRPC client expects just the hostname:port
	server := strings.TrimPrefix(cfg.Server, "https://")
	server = strings.TrimPrefix(server, "http://")

	// Create Argo CD client
	clientOpts := apiclient.ClientOptions{
		ServerAddr: server,
		AuthToken:  cfg.AuthToken,
		Insecure:   cfg.Insecure,
	}

	apiClient, err := apiclient.NewClient(&clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Argo CD client: %w", err)
	}

	return &ClientWithServer{
		Client: apiClient,
		Server: server,
	}, nil
}
