package argo

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListClustersInput defines the input parameters for listing Argo clusters
// Currently no additional parameters are needed as the client is configured at startup
type ListClustersInput struct {
}

// ListClustersOutput defines the output structure for listing Argo clusters
type ListClustersOutput struct {
	Items interface{} `json:"items" jsonschema:"raw cluster list from Argo CD API"`
}

// NewListClustersHandler creates a ListClusters handler with the provided ArgoCD client
func NewListClustersHandler(apiClient apiclient.Client) func(context.Context, *mcp.CallToolRequest, ListClustersInput) (*mcp.CallToolResult, ListClustersOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListClustersInput) (*mcp.CallToolResult, ListClustersOutput, error) {
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
}
