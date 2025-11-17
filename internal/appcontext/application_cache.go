package appcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"template_cli/internal/log"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

const (
	// ApplicationCacheFile is the filename for the application cache
	ApplicationCacheFile = "application_cache.json"

	// ApplicationCacheTTL is the default time-to-live for application cache
	ApplicationCacheTTL = 60 * time.Minute
)

// ApplicationCache represents cached application list data
type ApplicationCache struct {
	Items     []v1alpha1.Application `json:"items"`
	CachedAt  time.Time              `json:"cached_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// GetCachedApplications retrieves the cached application list if it's still valid
// Returns nil if cache is expired or doesn't exist
func (ac *AppContext) GetCachedApplications() *ApplicationCache {
	ac.applicationCacheMutex.RLock()
	defer ac.applicationCacheMutex.RUnlock()

	if ac.applicationCache == nil {
		return nil
	}

	if time.Now().After(ac.applicationCache.ExpiresAt) {
		return nil
	}

	return ac.applicationCache
}

// SetApplicationCache updates the application cache with the given items and TTL
func (ac *AppContext) SetApplicationCache(items []v1alpha1.Application, ttl time.Duration) {
	ac.applicationCacheMutex.Lock()
	defer ac.applicationCacheMutex.Unlock()

	now := time.Now()
	ac.applicationCache = &ApplicationCache{
		Items:     items,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
	}

	// Persist to disk
	if err := ac.writeApplicationCacheToDisk(); err != nil {
		log.Logger().Warnw("Failed to write application cache to disk", "error", err)
	}
}

// InvalidateApplicationCache clears the application cache
func (ac *AppContext) InvalidateApplicationCache() {
	ac.applicationCacheMutex.Lock()
	defer ac.applicationCacheMutex.Unlock()

	ac.applicationCache = nil

	// Remove cache file from disk
	cachePath := filepath.Join(log.ContextDir, ApplicationCacheFile)
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		log.Logger().Warnw("Failed to remove application cache file", "error", err)
	}
}

// RefreshApplicationCache fetches fresh application data from ArgoCD and caches it
// Returns error if the fetch fails
func (ac *AppContext) RefreshApplicationCache(ctxIn context.Context) error {
	l := log.Logger().With("component", "refresh_application_cache")

	l.Info("Fetching fresh application data from ArgoCD")
	conn, appClient, err := ac.ArgoClient.NewApplicationClient()
	if err != nil {
		l.Errorw("Failed to create application client", "error", err)
		return fmt.Errorf("failed to create application client: %w", err)
	}
	defer conn.Close()

	// List applications with timing
	listStartTime := time.Now()
	appList, err := appClient.List(ctxIn, &application.ApplicationQuery{})
	listDuration := time.Since(listStartTime)

	if err != nil {
		l.Errorw("Failed to list applications", "error", err, "duration", listDuration)
		return fmt.Errorf("failed to list applications: %w", err)
	}

	l.Infow("Successfully fetched applications from ArgoCD", "count", len(appList.Items), "duration", listDuration.String())

	// Cache the results
	ac.SetApplicationCache(appList.Items, ApplicationCacheTTL)

	return nil
}

// writeApplicationCacheToDisk persists the application cache to disk (caller must hold lock)
func (ac *AppContext) writeApplicationCacheToDisk() error {
	if ac.applicationCache == nil {
		return nil
	}

	cachePath := filepath.Join(log.ContextDir, ApplicationCacheFile)

	data, err := json.MarshalIndent(ac.applicationCache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal application cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write application cache file: %w", err)
	}

	return nil
}

// loadApplicationCacheFromDisk loads the application cache from disk if it exists and is valid
// If the cache is expired, it will be refreshed from ArgoCD
func (ac *AppContext) loadApplicationCacheFromDisk() {
	ac.applicationCacheMutex.Lock()
	cachePath := filepath.Join(log.ContextDir, ApplicationCacheFile)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Logger().Warnw("Failed to read application cache file", "error", err)
		}
		ac.applicationCacheMutex.Unlock()
		// No cache exists, attempt to refresh
		log.Logger().Info("No application cache found, fetching fresh data")
		if err := ac.RefreshApplicationCache(context.Background()); err != nil {
			log.Logger().Warnw("Failed to refresh application cache on startup", "error", err)
		}
		return
	}

	var cache ApplicationCache
	if err := json.Unmarshal(data, &cache); err != nil {
		log.Logger().Warnw("Failed to unmarshal application cache", "error", err)
		ac.applicationCacheMutex.Unlock()
		return
	}

	// Check if cache is expired
	if time.Now().After(cache.ExpiresAt) {
		// Cache is expired, remove the file
		os.Remove(cachePath)
		ac.applicationCacheMutex.Unlock()

		// Refresh the cache with fresh data
		log.Logger().Info("Application cache expired, fetching fresh data")
		if err := ac.RefreshApplicationCache(context.Background()); err != nil {
			log.Logger().Warnw("Failed to refresh expired application cache", "error", err)
		}
		return
	}

	// Cache is valid, use it
	ac.applicationCache = &cache
	ac.applicationCacheMutex.Unlock()
}
