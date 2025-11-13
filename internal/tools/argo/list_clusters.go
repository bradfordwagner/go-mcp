package argo

import (
	"context"
	"fmt"
	"time"

	"template_cli/internal/appcontext"
	"template_cli/internal/log"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

// ListClustersInput defines the input parameters for listing Argo clusters
// Currently no additional parameters are needed as the client is configured at startup
type ListClustersInput struct {
}

// ListClustersOutput defines the output structure for listing Argo clusters
type ListClustersOutput struct {
	Items interface{} `json:"items" jsonschema:"raw cluster list from Argo CD API"`
}

const (
	// ClusterCacheTTL is the default time-to-live for cluster cache
	ClusterCacheTTL = 60 * time.Minute
)

// NewListClustersHandler creates a ListClusters handler with the provided AppContext
func NewListClustersHandler(appCtx *appcontext.AppContext) func(context.Context, *mcp.CallToolRequest, ListClustersInput) (*mcp.CallToolResult, ListClustersOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListClustersInput) (*mcp.CallToolResult, ListClustersOutput, error) {
		// Check if we have cached clusters
		if cachedClusters := appCtx.GetCachedClusters(); cachedClusters != nil {
			log.Logger().Info("Returning cached clusters", zap.Int("count", len(cachedClusters.Items)))
			return nil, ListClustersOutput{
				Items: cachedClusters.Items,
			}, nil
		}

		// Cache miss or expired - fetch from ArgoCD
		log.Logger().Info("Cache miss, fetching clusters from ArgoCD")
		conn, clusterClient, err := appCtx.ArgoClient.NewClusterClient()
		if err != nil {
			log.Logger().Error("Failed to create cluster client", zap.Error(err))
			return nil, ListClustersOutput{}, fmt.Errorf("failed to create cluster client: %w", err)
		}
		defer conn.Close()

		// List clusters
		clusterList, err := clusterClient.List(ctx, &cluster.ClusterQuery{})
		if err != nil {
			log.Logger().Error("Failed to list clusters", zap.Error(err))
			return nil, ListClustersOutput{}, fmt.Errorf("failed to list clusters: %w", err)
		}

		log.Logger().Info("Successfully fetched clusters from ArgoCD", zap.Int("count", len(clusterList.Items)))

		// Cache the results
		appCtx.SetClusterCache(clusterList.Items, ClusterCacheTTL)

		// Return raw response from Argo CD API
		return nil, ListClustersOutput{
			Items: clusterList.Items,
		}, nil
	}
}
