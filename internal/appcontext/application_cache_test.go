package appcontext

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ApplicationCache", func() {
	var ac *AppContext

	BeforeEach(func() {
		ac = &AppContext{
			applicationCacheMutex: sync.RWMutex{},
		}
	})

	Describe("GetCachedApplications", func() {
		Context("when cache is nil", func() {
			It("should return nil", func() {
				result := ac.GetCachedApplications()
				Expect(result).To(BeNil())
			})
		})

		Context("when cache is valid", func() {
			BeforeEach(func() {
				ac.applicationCache = &ApplicationCache{
					Items:     createTestApps(2),
					CachedAt:  time.Now().Add(-30 * time.Minute),
					ExpiresAt: time.Now().Add(30 * time.Minute),
				}
			})

			It("should return the cached items", func() {
				result := ac.GetCachedApplications()
				Expect(result).NotTo(BeNil())
				Expect(result.Items).To(HaveLen(2))
			})
		})

		Context("when cache is expired", func() {
			BeforeEach(func() {
				ac.applicationCache = &ApplicationCache{
					Items:     createTestApps(2),
					CachedAt:  time.Now().Add(-2 * time.Hour),
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}
			})

			It("should return nil", func() {
				result := ac.GetCachedApplications()
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("SetApplicationCache", func() {
		DescribeTable("should set cache with different TTLs",
			func(count int, ttl time.Duration) {
				apps := createTestApps(count)
				beforeSet := time.Now()
				ac.SetApplicationCache(apps, ttl)
				afterSet := time.Now()

				Expect(ac.applicationCache).NotTo(BeNil())
				Expect(ac.applicationCache.Items).To(HaveLen(count))
				Expect(ac.applicationCache.CachedAt).To(BeTemporally(">=", beforeSet))
				Expect(ac.applicationCache.CachedAt).To(BeTemporally("<=", afterSet))

				actualTTL := ac.applicationCache.ExpiresAt.Sub(ac.applicationCache.CachedAt)
				Expect(actualTTL).To(BeNumerically("~", ttl, time.Second))
			},
			Entry("1 hour TTL with 3 apps", 3, 1*time.Hour),
			Entry("30 minute TTL with 1 app", 1, 30*time.Minute),
			Entry("2 hour TTL with 5 apps", 5, 2*time.Hour),
		)
	})

	Describe("InvalidateApplicationCache", func() {
		BeforeEach(func() {
			ac.applicationCache = &ApplicationCache{
				Items:     createTestApps(2),
				CachedAt:  time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}
		})

		It("should clear the cache", func() {
			ac.InvalidateApplicationCache()
			Expect(ac.applicationCache).To(BeNil())
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
					apps := createTestApps(n % 5)
					ac.SetApplicationCache(apps, 1*time.Hour)
				}(i)
			}

			// Concurrent reads
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = ac.GetCachedApplications()
				}()
			}

			// Concurrent invalidations
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ac.InvalidateApplicationCache()
				}()
			}

			wg.Wait()
		})
	})

	Describe("Cache Expiration", func() {
		It("should expire after TTL", func() {
			apps := createTestApps(1)
			ac.SetApplicationCache(apps, 100*time.Millisecond)

			// Should be valid immediately
			result := ac.GetCachedApplications()
			Expect(result).NotTo(BeNil())

			// Wait for expiration
			time.Sleep(150 * time.Millisecond)

			// Should be expired now
			result = ac.GetCachedApplications()
			Expect(result).To(BeNil())
		})
	})

	Describe("Empty Items", func() {
		It("should handle empty application list", func() {
			apps := []v1alpha1.Application{}
			ac.SetApplicationCache(apps, 1*time.Hour)

			result := ac.GetCachedApplications()
			Expect(result).NotTo(BeNil())
			Expect(result.Items).To(BeEmpty())
		})
	})

	Describe("writeApplicationCacheToDisk", func() {
		Context("when cache is nil", func() {
			It("should not return an error", func() {
				ac.applicationCache = nil
				err := ac.writeApplicationCacheToDisk()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

// Helper function to create test applications
func createTestApps(count int) []v1alpha1.Application {
	apps := make([]v1alpha1.Application, count)
	for i := 0; i < count; i++ {
		apps[i] = v1alpha1.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-app",
				Namespace: "default",
			},
			Spec: v1alpha1.ApplicationSpec{
				Project: "default",
			},
		}
	}
	return apps
}
