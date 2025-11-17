package argo

import (
	"context"
	"fmt"
	"time"

	"template_cli/internal/appcontext"
	"template_cli/internal/log"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListApplicationsInput defines the input parameters for listing Argo applications
type ListApplicationsInput struct {
	Project   string `json:"project,omitempty" jsonschema:"optional project filter"`
	Namespace string `json:"namespace,omitempty" jsonschema:"optional namespace filter"`
	Cluster   string `json:"cluster,omitempty" jsonschema:"optional cluster filter"`
}

// ListApplicationsOutput defines the output structure for listing Argo applications
type ListApplicationsOutput struct {
	Items interface{} `json:"items" jsonschema:"raw application list from Argo CD API"`
}

// NewListApplicationsHandler creates a ListApplications handler with the provided AppContext
func NewListApplicationsHandler(appCtx *appcontext.AppContext) func(context.Context, *mcp.CallToolRequest, ListApplicationsInput) (*mcp.CallToolResult, ListApplicationsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListApplicationsInput) (*mcp.CallToolResult, ListApplicationsOutput, error) {
		l := log.Logger().With("component", "argocd_list_applications")
		startTime := time.Now()
		defer func() {
			duration := time.Since(startTime)
			l.Infow("list_applications completed", "duration", duration)
		}()

		// Check if we have cached applications
		if cachedApps := appCtx.GetCachedApplications(); cachedApps != nil {
			l.Infow("Returning cached applications", "count", len(cachedApps.Items))
			
			// Apply filters if provided
			filteredApps := filterApplications(cachedApps.Items, input)
			l.Infow("Filtered applications", "filtered_count", len(filteredApps))
			
			return nil, ListApplicationsOutput{
				Items: filteredApps,
			}, nil
		}

		// Cache miss or expired - refresh from ArgoCD
		l.Info("Cache miss, refreshing applications from ArgoCD")
		if err := appCtx.RefreshApplicationCache(ctx); err != nil {
			return nil, ListApplicationsOutput{}, fmt.Errorf("failed to refresh application cache: %w", err)
		}

		// Get the freshly cached applications
		cachedApps := appCtx.GetCachedApplications()
		if cachedApps == nil {
			return nil, ListApplicationsOutput{}, fmt.Errorf("application cache is unexpectedly empty after refresh")
		}

		// Apply filters if provided
		filteredApps := filterApplications(cachedApps.Items, input)
		l.Infow("Filtered applications", "filtered_count", len(filteredApps))

		// Return filtered response
		return nil, ListApplicationsOutput{
			Items: filteredApps,
		}, nil
	}
}

// filterApplications applies the optional filters to the application list
func filterApplications(apps []v1alpha1.Application, input ListApplicationsInput) []v1alpha1.Application {
	// If no filters are provided, return all applications
	if input.Project == "" && input.Namespace == "" && input.Cluster == "" {
		return apps
	}

	filtered := make([]v1alpha1.Application, 0)
	for _, app := range apps {
		// Check project filter
		if input.Project != "" && app.Spec.Project != input.Project {
			continue
		}

		// Check namespace filter
		if input.Namespace != "" && app.Namespace != input.Namespace {
			continue
		}

		// Check cluster filter
		if input.Cluster != "" && app.Spec.Destination.Server != input.Cluster {
			continue
		}

		// Application matches all filters
		filtered = append(filtered, app)
	}

	return filtered
}

