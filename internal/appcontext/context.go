package appcontext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

const (
	// ContextDir is the directory where cache files are stored
	ContextDir = "/tmp/bw-mcp"

	// ClusterCacheFile is the filename for the cluster cache
	ClusterCacheFile = "cluster_cache.json"
)

// AppContext holds shared application state and dependencies
type AppContext struct {
	// ArgoClient is the initialized ArgoCD API client
	ArgoClient apiclient.Client

	// ClusterCache holds cached cluster information
	clusterCache      *ClusterCache
	clusterCacheMutex sync.RWMutex
}

// ClusterCache represents cached cluster list data
type ClusterCache struct {
	Items     []v1alpha1.Cluster `json:"items"`
	CachedAt  time.Time          `json:"cached_at"`
	ExpiresAt time.Time          `json:"expires_at"`
}

// NewAppContext creates a new application context
func NewAppContext(argoClient apiclient.Client) *AppContext {
	ctx := &AppContext{
		ArgoClient: argoClient,
	}

	// Ensure context directory exists
	if err := os.MkdirAll(ContextDir, 0755); err != nil {
		// Log error but don't fail - we can still run without cache
		fmt.Fprintf(os.Stderr, "Warning: failed to create context directory: %v\n", err)
	}

	// Try to load existing cache from disk
	ctx.loadClusterCacheFromDisk()

	return ctx
}

// GetCachedClusters retrieves the cached cluster list if it's still valid
// Returns nil if cache is expired or doesn't exist
func (ctx *AppContext) GetCachedClusters() *ClusterCache {
	ctx.clusterCacheMutex.RLock()
	defer ctx.clusterCacheMutex.RUnlock()

	if ctx.clusterCache == nil {
		return nil
	}

	if time.Now().After(ctx.clusterCache.ExpiresAt) {
		return nil
	}

	return ctx.clusterCache
}

// SetClusterCache updates the cluster cache with the given items and TTL
func (ctx *AppContext) SetClusterCache(items []v1alpha1.Cluster, ttl time.Duration) {
	ctx.clusterCacheMutex.Lock()
	defer ctx.clusterCacheMutex.Unlock()

	now := time.Now()
	ctx.clusterCache = &ClusterCache{
		Items:     items,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}

	// Persist to disk
	if err := ctx.writeClusterCacheToDisk(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write cluster cache to disk: %v\n", err)
	}
}

// InvalidateClusterCache clears the cluster cache
func (ctx *AppContext) InvalidateClusterCache() {
	ctx.clusterCacheMutex.Lock()
	defer ctx.clusterCacheMutex.Unlock()

	ctx.clusterCache = nil

	// Remove cache file from disk
	cachePath := filepath.Join(ContextDir, ClusterCacheFile)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove cluster cache file: %v\n", err)
	}
}

// writeClusterCacheToDisk persists the cluster cache to disk (caller must hold lock)
func (ctx *AppContext) writeClusterCacheToDisk() error {
	if ctx.clusterCache == nil {
		return nil
	}

	cachePath := filepath.Join(ContextDir, ClusterCacheFile)

	data, err := json.MarshalIndent(ctx.clusterCache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cluster cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cluster cache file: %w", err)
	}

	return nil
}

// loadClusterCacheFromDisk loads the cluster cache from disk if it exists and is valid
func (ctx *AppContext) loadClusterCacheFromDisk() {
	ctx.clusterCacheMutex.Lock()
	defer ctx.clusterCacheMutex.Unlock()

	cachePath := filepath.Join(ContextDir, ClusterCacheFile)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to read cluster cache file: %v\n", err)
		}
		return
	}

	var cache ClusterCache
	if err := json.Unmarshal(data, &cache); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to unmarshal cluster cache: %v\n", err)
		return
	}

	// Check if cache is expired
	if time.Now().After(cache.ExpiresAt) {
		// Cache is expired, remove the file
		os.Remove(cachePath)
		return
	}

	// Cache is valid, use it
	ctx.clusterCache = &cache
}
