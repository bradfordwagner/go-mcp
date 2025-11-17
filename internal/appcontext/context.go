package appcontext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"template_cli/internal/log"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
)

const (
	// ServerConfigFile is the filename for the server configuration cache
	ServerConfigFile = "server_config.json"
)

// AppContext holds shared application state and dependencies
type AppContext struct {
	// ArgoClient is the initialized ArgoCD API client
	ArgoClient apiclient.Client

	// ArgoServer is the ArgoCD server URL we're connected to
	ArgoServer string

	// ClusterCache holds cached cluster information
	clusterCache      *ClusterCache
	clusterCacheMutex sync.RWMutex

	// ApplicationCache holds cached application information
	applicationCache      *ApplicationCache
	applicationCacheMutex sync.RWMutex
}

// ServerConfig represents cached server configuration
type ServerConfig struct {
	Server  string    `json:"server"`
	SavedAt time.Time `json:"saved_at"`
}

// NewAppContext creates a new application context
// If the server URL has changed since the last run, all caches will be invalidated
func NewAppContext(argoClient apiclient.Client, argoServer string) *AppContext {
	ctx := &AppContext{
		ArgoClient: argoClient,
		ArgoServer: argoServer,
	}

	// Ensure context directory exists
	if err := os.MkdirAll(log.ContextDir, 0755); err != nil {
		// Log error but don't fail - we can still run without cache
		log.Logger().Warnw("Failed to create context directory", "error", err)
	}

	// Check if server has changed and invalidate caches if needed
	if ctx.hasServerChanged() {
		log.Logger().Info("ArgoCD server has changed, invalidating all caches")
		ctx.deleteAllCaches()
	}

	// Save current server configuration
	ctx.saveServerConfig()

	// Try to load existing caches from disk
	ctx.loadClusterCacheFromDisk()
	ctx.loadApplicationCacheFromDisk()

	return ctx
}

// hasServerChanged checks if the current server URL differs from the cached one
func (ctx *AppContext) hasServerChanged() bool {
	serverConfigPath := filepath.Join(log.ContextDir, ServerConfigFile)

	data, err := os.ReadFile(serverConfigPath)
	if err != nil {
		// If file doesn't exist, this is first run or cache was cleared
		if os.IsNotExist(err) {
			return false
		}
		log.Logger().Warnw("Failed to read server config file", "error", err)
		return false
	}

	var serverConfig ServerConfig
	if err := json.Unmarshal(data, &serverConfig); err != nil {
		log.Logger().Warnw("Failed to unmarshal server config", "error", err)
		return false
	}

	// Compare cached server with current server
	return serverConfig.Server != ctx.ArgoServer
}

// saveServerConfig saves the current server configuration to disk
func (ctx *AppContext) saveServerConfig() {
	serverConfigPath := filepath.Join(log.ContextDir, ServerConfigFile)

	serverConfig := ServerConfig{
		Server:  ctx.ArgoServer,
		SavedAt: time.Now(),
	}

	data, err := json.MarshalIndent(serverConfig, "", "  ")
	if err != nil {
		log.Logger().Warnw("Failed to marshal server config", "error", err)
		return
	}

	if err := os.WriteFile(serverConfigPath, data, 0644); err != nil {
		log.Logger().Warnw("Failed to write server config file", "error", err)
	}
}

// deleteAllCaches removes all cache files from disk
func (ctx *AppContext) deleteAllCaches() {
	// List of cache files to delete
	cacheFiles := []string{
		ClusterCacheFile,
		ApplicationCacheFile,
		// Add more cache files here as they are added to the system
	}

	for _, cacheFile := range cacheFiles {
		cachePath := filepath.Join(log.ContextDir, cacheFile)
		if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
			log.Logger().Warnw("Failed to remove cache file", "file", cacheFile, "error", err)
		}
	}

	// Clear in-memory caches
	ctx.clusterCacheMutex.Lock()
	ctx.clusterCache = nil
	ctx.clusterCacheMutex.Unlock()

	ctx.applicationCacheMutex.Lock()
	ctx.applicationCache = nil
	ctx.applicationCacheMutex.Unlock()
}
