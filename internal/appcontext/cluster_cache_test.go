package appcontext

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

var _ = Describe("ClusterCache", func() {
	var ctx *AppContext

	BeforeEach(func() {
		ctx = &AppContext{
			clusterCacheMutex: sync.RWMutex{},
		}
	})

	Describe("GetCachedClusters", func() {
		Context("when cache is nil", func() {
			It("should return nil", func() {
				result := ctx.GetCachedClusters()
				Expect(result).To(BeNil())
			})
		})

		Context("when cache is valid", func() {
			BeforeEach(func() {
				ctx.clusterCache = &ClusterCache{
					Items:     createTestClusters(3),
					CachedAt:  time.Now().Add(-30 * time.Minute),
					ExpiresAt: time.Now().Add(30 * time.Minute),
				}
			})

			It("should return the cached items", func() {
				result := ctx.GetCachedClusters()
				Expect(result).NotTo(BeNil())
				Expect(result.Items).To(HaveLen(3))
			})
		})

		Context("when cache is expired", func() {
			BeforeEach(func() {
				ctx.clusterCache = &ClusterCache{
					Items:     createTestClusters(2),
					CachedAt:  time.Now().Add(-2 * time.Hour),
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}
			})

			It("should return nil", func() {
				result := ctx.GetCachedClusters()
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("SetClusterCache", func() {
		DescribeTable("should set cache with different TTLs",
			func(count int, ttl time.Duration) {
				clusters := createTestClusters(count)
				beforeSet := time.Now()
				ctx.SetClusterCache(clusters, ttl)
				afterSet := time.Now()

				Expect(ctx.clusterCache).NotTo(BeNil())
				Expect(ctx.clusterCache.Items).To(HaveLen(count))
				Expect(ctx.clusterCache.CachedAt).To(BeTemporally(">=", beforeSet))
				Expect(ctx.clusterCache.CachedAt).To(BeTemporally("<=", afterSet))

				actualTTL := ctx.clusterCache.ExpiresAt.Sub(ctx.clusterCache.CachedAt)
				Expect(actualTTL).To(BeNumerically("~", ttl, time.Second))
			},
			Entry("1 hour TTL with 2 clusters", 2, 1*time.Hour),
			Entry("30 minute TTL with 1 cluster", 1, 30*time.Minute),
			Entry("45 minute TTL with 5 clusters", 5, 45*time.Minute),
		)
	})

	Describe("InvalidateClusterCache", func() {
		BeforeEach(func() {
			ctx.clusterCache = &ClusterCache{
				Items:     createTestClusters(2),
				CachedAt:  time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}
		})

		It("should clear the cache", func() {
			ctx.InvalidateClusterCache()
			Expect(ctx.clusterCache).To(BeNil())
		})
	})

	Describe("Cache Concurrency", func() {
		It("should handle concurrent operations safely", func() {
			var wg sync.WaitGroup
			iterations := 100

			// Concurrent writes
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func(n int) {
					defer wg.Done()
					clusters := createTestClusters(n % 5)
					ctx.SetClusterCache(clusters, 1*time.Hour)
				}(i)
			}

			// Concurrent reads
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = ctx.GetCachedClusters()
				}()
			}

			// Concurrent invalidations
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ctx.InvalidateClusterCache()
				}()
			}

			wg.Wait()
		})
	})

	Describe("Cache Expiration", func() {
		It("should expire after TTL", func() {
			clusters := createTestClusters(1)
			ctx.SetClusterCache(clusters, 100*time.Millisecond)

			// Should be valid immediately
			result := ctx.GetCachedClusters()
			Expect(result).NotTo(BeNil())

			// Wait for expiration
			time.Sleep(150 * time.Millisecond)

			// Should be expired now
			result = ctx.GetCachedClusters()
			Expect(result).To(BeNil())
		})
	})

	Describe("Empty Items", func() {
		It("should handle empty cluster list", func() {
			clusters := []v1alpha1.Cluster{}
			ctx.SetClusterCache(clusters, 1*time.Hour)

			result := ctx.GetCachedClusters()
			Expect(result).NotTo(BeNil())
			Expect(result.Items).To(BeEmpty())
		})
	})

	Describe("Cache Constants", func() {
		It("should have correct TTL constant", func() {
			Expect(ClusterCacheTTL).To(Equal(60 * time.Minute))
		})

		It("should have correct cache file name", func() {
			Expect(ClusterCacheFile).To(Equal("cluster_cache.json"))
		})
	})

	Describe("writeClusterCacheToDisk", func() {
		Context("when cache is nil", func() {
			It("should not return an error", func() {
				ctx.clusterCache = nil
				err := ctx.writeClusterCacheToDisk()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Multiple Updates", func() {
		It("should replace previous cache", func() {
			// First update
			clusters1 := createTestClusters(2)
			ctx.SetClusterCache(clusters1, 1*time.Hour)

			result1 := ctx.GetCachedClusters()
			Expect(result1).NotTo(BeNil())
			Expect(result1.Items).To(HaveLen(2))

			// Second update (should replace)
			clusters2 := createTestClusters(5)
			ctx.SetClusterCache(clusters2, 30*time.Minute)

			result2 := ctx.GetCachedClusters()
			Expect(result2).NotTo(BeNil())
			Expect(result2.Items).To(HaveLen(5))
		})
	})
})

// Helper function to create test clusters
func createTestClusters(count int) []v1alpha1.Cluster {
	clusters := make([]v1alpha1.Cluster, count)
	for i := 0; i < count; i++ {
		clusters[i] = v1alpha1.Cluster{
			Name:   "test-cluster",
			Server: "https://kubernetes.default.svc",
		}
	}
	return clusters
}
