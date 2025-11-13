package appcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"template_cli/internal/log"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

const (
	// ClusterCacheFile is the filename for the cluster cache
	ClusterCacheFile = "cluster_cache.json"

	// ClusterCacheTTL is the default time-to-live for cluster cache
	ClusterCacheTTL = 60 * time.Minute
)

// ClusterCache represents cached cluster list data
type ClusterCache struct {
	Items     []v1alpha1.Cluster `json:"items"`
	CachedAt  time.Time          `json:"cached_at"`
	ExpiresAt time.Time          `json:"expires_at"`
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
		log.Logger().Warnw("Failed to write cluster cache to disk", "error", err)
	}
}

// InvalidateClusterCache clears the cluster cache
func (ctx *AppContext) InvalidateClusterCache() {
	ctx.clusterCacheMutex.Lock()
	defer ctx.clusterCacheMutex.Unlock()

	ctx.clusterCache = nil

	// Remove cache file from disk
	cachePath := filepath.Join(log.ContextDir, ClusterCacheFile)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		log.Logger().Warnw("Failed to remove cluster cache file", "error", err)
	}
}

// RefreshClusterCache fetches fresh cluster data from ArgoCD and caches it
// Returns error if the fetch fails
func (ctx *AppContext) RefreshClusterCache(ctxIn context.Context) error {
	l := log.Logger().With("component", "refresh_cluster_cache")

	l.Info("Fetching fresh cluster data from ArgoCD")
	conn, clusterClient, err := ctx.ArgoClient.NewClusterClient()
	if err != nil {
		l.Errorw("Failed to create cluster client", "error", err)
		return fmt.Errorf("failed to create cluster client: %w", err)
	}
	defer conn.Close()

	// List clusters
	clusterList, err := clusterClient.List(ctxIn, &cluster.ClusterQuery{})
	if err != nil {
		l.Errorw("Failed to list clusters", "error", err)
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	l.Infow("Successfully fetched clusters from ArgoCD", "count", len(clusterList.Items))

	// Cache the results
	ctx.SetClusterCache(clusterList.Items, ClusterCacheTTL)

	return nil
}

// writeClusterCacheToDisk persists the cluster cache to disk (caller must hold lock)
func (ctx *AppContext) writeClusterCacheToDisk() error {
	if ctx.clusterCache == nil {
		return nil
	}

	cachePath := filepath.Join(log.ContextDir, ClusterCacheFile)

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
// If the cache is expired, it will be refreshed from ArgoCD
func (ctx *AppContext) loadClusterCacheFromDisk() {
	ctx.clusterCacheMutex.Lock()
	cachePath := filepath.Join(log.ContextDir, ClusterCacheFile)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Logger().Warnw("Failed to read cluster cache file", "error", err)
		}
		ctx.clusterCacheMutex.Unlock()
		// No cache exists, attempt to refresh
		log.Logger().Info("No cluster cache found, fetching fresh data")
		if err := ctx.RefreshClusterCache(context.Background()); err != nil {
			log.Logger().Warnw("Failed to refresh cluster cache on startup", "error", err)
		}
		return
	}

	var cache ClusterCache
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Logger().Warnw("Failed to unmarshal cluster cache", "error", err)
		ctx.clusterCacheMutex.Unlock()
		return
	}

	// Check if cache is expired
	if time.Now().After(cache.ExpiresAt) {
		// Cache is expired, remove the file
		os.Remove(cachePath)
		ctx.clusterCacheMutex.Unlock()

		// Refresh the cache with fresh data
		log.Logger().Info("Cluster cache expired, fetching fresh data")
		if err := ctx.RefreshClusterCache(context.Background()); err != nil {
			log.Logger().Warnw("Failed to refresh expired cluster cache", "error", err)
		}
		return
	}

	// Cache is valid, use it
	ctx.clusterCache = &cache
	ctx.clusterCacheMutex.Unlock()
}
