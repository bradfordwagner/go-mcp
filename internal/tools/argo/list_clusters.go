package argo

import (
	"context"
	"fmt"
	"time"

	"template_cli/internal/appcontext"
	"template_cli/internal/log"

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

const (
	// ClusterCacheTTL is the default time-to-live for cluster cache
	ClusterCacheTTL = 60 * time.Minute
)

// NewListClustersHandler creates a ListClusters handler with the provided AppContext
func NewListClustersHandler(appCtx *appcontext.AppContext) func(context.Context, *mcp.CallToolRequest, ListClustersInput) (*mcp.CallToolResult, ListClustersOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListClustersInput) (*mcp.CallToolResult, ListClustersOutput, error) {
		l := log.Logger().With("component", "argocd_list_clusters")
		startTime := time.Now()
		defer func() {
			duration := time.Since(startTime)
			l.Infow("list_clusters completed", "duration", duration)
		}()

		// Check if we have cached clusters
		if cachedClusters := appCtx.GetCachedClusters(); cachedClusters != nil {
			l.Infow("Returning cached clusters", "count", len(cachedClusters.Items))
			return nil, ListClustersOutput{
				Items: cachedClusters.Items,
			}, nil
		}

		// Cache miss or expired - fetch from ArgoCD
		l.Info("Cache miss, fetching clusters from ArgoCD")
		conn, clusterClient, err := appCtx.ArgoClient.NewClusterClient()
		if err != nil {
			l.Errorw("Failed to create cluster client", "error", err)
			return nil, ListClustersOutput{}, fmt.Errorf("failed to create cluster client: %w", err)
		}
		defer conn.Close()

		// List clusters
		clusterList, err := clusterClient.List(ctx, &cluster.ClusterQuery{})
		if err != nil {
			l.Errorw("Failed to list clusters", "error", err)
			return nil, ListClustersOutput{}, fmt.Errorf("failed to list clusters: %w", err)
		}

		l.Infow("Successfully fetched clusters from ArgoCD", "count", len(clusterList.Items))

		// Cache the results
		appCtx.SetClusterCache(clusterList.Items, ClusterCacheTTL)

		// Return raw response from Argo CD API
		return nil, ListClustersOutput{
			Items: clusterList.Items,
		}, nil
	}
}
