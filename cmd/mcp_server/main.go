package main

import (
	"context"
	stdlog "log"

	"template_cli/internal/appcontext"
	"template_cli/internal/argoclient"
	"template_cli/internal/log"
	"template_cli/internal/tools/argo"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

type Input struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type Output struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input Input) (
	*mcp.CallToolResult,
	Output,
	error,
) {
	return nil, Output{Greeting: "Hi " + input.Name}, nil
}

func main() {
	// Initialize logger
	if err := log.Init(); err != nil {
		stdlog.Fatalf("Failed to initialize logger: %v", err)
	}
	defer log.Sync()

	log.Logger().Info("Starting MCP server")

	// Initialize ArgoCD client
	// Client config will be read from environment variables (ARGOCD_BASE_URL, ARGOCD_API_TOKEN)
	argoClientWithServer, err := argoclient.NewClient(argoclient.Config{})
	if err != nil {
		log.Logger().Fatal("Failed to create ArgoCD client", zap.Error(err))
	}

	// Create application context with shared state and dependencies
	// The server URL is passed to enable cache invalidation when it changes
	appCtx := appcontext.NewAppContext(argoClientWithServer.Client, argoClientWithServer.Server)

	// Create a server with multiple tools.
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)
	mcp.AddTool(server, &mcp.Tool{Name: "argocd_list_clusters", Description: "list Argo CD clusters"}, argo.NewListClustersHandler(appCtx))
	
	log.Logger().Info("MCP server initialized, starting server loop")
	
	// Run the server over stdin/stdout, until the client disconnects.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Logger().Fatal("Server error", zap.Error(err))
	}
}
