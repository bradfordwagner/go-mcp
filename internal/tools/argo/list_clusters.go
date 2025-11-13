package argo

import (
	"context"
	"fmt"
	"time"

	"template_cli/internal/appcontext"
	"template_cli/internal/log"

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

		// Cache miss or expired - refresh from ArgoCD
		l.Info("Cache miss, refreshing clusters from ArgoCD")
		if err := appCtx.RefreshClusterCache(ctx); err != nil {
			return nil, ListClustersOutput{}, fmt.Errorf("failed to refresh cluster cache: %w", err)
		}

		// Get the freshly cached clusters
		cachedClusters := appCtx.GetCachedClusters()
		if cachedClusters == nil {
			return nil, ListClustersOutput{}, fmt.Errorf("cluster cache is unexpectedly empty after refresh")
		}

		// Return cached response
		return nil, ListClustersOutput{
			Items: cachedClusters.Items,
		}, nil
	}
}
