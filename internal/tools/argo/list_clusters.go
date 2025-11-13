package argo

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListClustersInput defines the input parameters for listing Argo clusters
type ListClustersInput struct {
	Server    string `json:"server,omitempty" jsonschema:"Argo CD server address (defaults to ARGOCD_BASE_URL env var)"`
	AuthToken string `json:"auth_token,omitempty" jsonschema:"Argo CD auth token (defaults to ARGOCD_API_TOKEN env var)"`
	Insecure  bool   `json:"insecure,omitempty" jsonschema:"Skip TLS certificate verification (defaults to false)"`
}

// ListClustersOutput defines the output structure for listing Argo clusters
type ListClustersOutput struct {
	Items interface{} `json:"items" jsonschema:"raw cluster list from Argo CD API"`
}

// ListClusters retrieves a list of Argo CD clusters
func ListClusters(ctx context.Context, req *mcp.CallToolRequest, input ListClustersInput) (
	*mcp.CallToolResult,
	ListClustersOutput,
	error,
) {
	// Get server address from input or environment
	server := input.Server
	if server == "" {
		server = os.Getenv("ARGOCD_BASE_URL")
		if server == "" {
			return nil, ListClustersOutput{}, fmt.Errorf("argo CD server address not provided. Set ARGOCD_BASE_URL env var or pass server in input")
		}
	}

	// Strip URL scheme if present (https:// or http://)
	// The Argo CD gRPC client expects just the hostname:port
	server = strings.TrimPrefix(server, "https://")
	server = strings.TrimPrefix(server, "http://")

	// Get auth token from input or environment
	authToken := input.AuthToken
	if authToken == "" {
		authToken = os.Getenv("ARGOCD_API_TOKEN")
		if authToken == "" {
			return nil, ListClustersOutput{}, fmt.Errorf("argo CD auth token not provided. Set ARGOCD_API_TOKEN env var or pass auth_token in input")
		}
	}

	// Create Argo CD client
	clientOpts := apiclient.ClientOptions{
		ServerAddr: server,
		AuthToken:  authToken,
		Insecure:   input.Insecure,
	}

	apiClient, err := apiclient.NewClient(&clientOpts)
	if err != nil {
		return nil, ListClustersOutput{}, fmt.Errorf("failed to create Argo CD client: %w", err)
	}

	// Get cluster client
	conn, clusterClient, err := apiClient.NewClusterClient()
	if err != nil {
		return nil, ListClustersOutput{}, fmt.Errorf("failed to create cluster client: %w", err)
	}
	defer conn.Close()

	// List clusters
	clusterList, err := clusterClient.List(ctx, &cluster.ClusterQuery{})
	if err != nil {
		return nil, ListClustersOutput{}, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Return raw response from Argo CD API
	return nil, ListClustersOutput{
		Items: clusterList.Items,
	}, nil
}
