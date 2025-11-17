package appcontext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"template_cli/internal/log"
)

var _ = Describe("AppContext", func() {
	Describe("ServerConfig Structure", func() {
		It("should store server and timestamp", func() {
			now := time.Now()
			config := ServerConfig{
				Server:  "test-server:443",
				SavedAt: now,
			}

			Expect(config.Server).To(Equal("test-server:443"))
			Expect(config.SavedAt).To(Equal(now))
		})
	})

	Describe("AppContext Structure", func() {
		It("should initialize with correct defaults", func() {
			ctx := &AppContext{
				ArgoClient: nil,
				ArgoServer: "test-server:443",
			}

			Expect(ctx.ArgoServer).To(Equal("test-server:443"))
			Expect(ctx.clusterCache).To(BeNil())
			Expect(ctx.applicationCache).To(BeNil())
		})
	})

	Describe("deleteAllCaches", func() {
		var ctx *AppContext

		BeforeEach(func() {
			// Ensure context directory exists
			Expect(os.MkdirAll(log.ContextDir, 0755)).To(Succeed())

			ctx = &AppContext{
				clusterCacheMutex:     sync.RWMutex{},
				applicationCacheMutex: sync.RWMutex{},
			}

			// Set some in-memory caches
			ctx.clusterCache = &ClusterCache{
				Items:     createTestClusters(1),
				CachedAt:  time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}

			ctx.applicationCache = &ApplicationCache{
				Items:     createTestApps(1),
				CachedAt:  time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}

			// Create cache files on disk
			clusterCachePath := filepath.Join(log.ContextDir, ClusterCacheFile)
			appCachePath := filepath.Join(log.ContextDir, ApplicationCacheFile)

			Expect(os.WriteFile(clusterCachePath, []byte("{}"), 0644)).To(Succeed())
			Expect(os.WriteFile(appCachePath, []byte("{}"), 0644)).To(Succeed())
		})

		It("should clear in-memory caches", func() {
			ctx.deleteAllCaches()

			Expect(ctx.clusterCache).To(BeNil())
			Expect(ctx.applicationCache).To(BeNil())
		})

		It("should delete cache files from disk", func() {
			clusterCachePath := filepath.Join(log.ContextDir, ClusterCacheFile)
			appCachePath := filepath.Join(log.ContextDir, ApplicationCacheFile)

			ctx.deleteAllCaches()

			Expect(clusterCachePath).NotTo(BeAnExistingFile())
			Expect(appCachePath).NotTo(BeAnExistingFile())
		})

		Context("when cache files don't exist", func() {
			BeforeEach(func() {
				os.Remove(filepath.Join(log.ContextDir, ClusterCacheFile))
				os.Remove(filepath.Join(log.ContextDir, ApplicationCacheFile))
			})

			It("should not error", func() {
				Expect(func() { ctx.deleteAllCaches() }).NotTo(Panic())
				Expect(ctx.clusterCache).To(BeNil())
				Expect(ctx.applicationCache).To(BeNil())
			})
		})
	})

	Describe("saveServerConfig", func() {
		var ctx *AppContext
		var serverConfigPath string

		BeforeEach(func() {
			Expect(os.MkdirAll(log.ContextDir, 0755)).To(Succeed())
			ctx = &AppContext{
				ArgoServer: "test-server:8080",
			}
			serverConfigPath = filepath.Join(log.ContextDir, ServerConfigFile)
		})

		AfterEach(func() {
			os.Remove(serverConfigPath)
		})

		It("should create server config file", func() {
			ctx.saveServerConfig()
			Expect(serverConfigPath).To(BeAnExistingFile())
		})

		It("should save correct server URL", func() {
			ctx.saveServerConfig()

			data, err := os.ReadFile(serverConfigPath)
			Expect(err).NotTo(HaveOccurred())

			var config ServerConfig
			Expect(json.Unmarshal(data, &config)).To(Succeed())
			Expect(config.Server).To(Equal("test-server:8080"))
		})

		It("should have recent timestamp", func() {
			ctx.saveServerConfig()

			data, err := os.ReadFile(serverConfigPath)
			Expect(err).NotTo(HaveOccurred())

			var config ServerConfig
			Expect(json.Unmarshal(data, &config)).To(Succeed())
			Expect(time.Since(config.SavedAt)).To(BeNumerically("<", 5*time.Second))
		})

		It("should update config on multiple saves", func() {
			ctx.saveServerConfig()

			data1, _ := os.ReadFile(serverConfigPath)
			var config1 ServerConfig
			json.Unmarshal(data1, &config1)

			time.Sleep(10 * time.Millisecond)

			ctx.ArgoServer = "server2-test:443"
			ctx.saveServerConfig()

			data2, _ := os.ReadFile(serverConfigPath)
			var config2 ServerConfig
			json.Unmarshal(data2, &config2)

			Expect(config2.Server).To(Equal("server2-test:443"))
			Expect(config2.SavedAt).To(BeTemporally(">", config1.SavedAt))
		})
	})

	Describe("hasServerChanged", func() {
		var ctx *AppContext
		var serverConfigPath string

		BeforeEach(func() {
			Expect(os.MkdirAll(log.ContextDir, 0755)).To(Succeed())
			serverConfigPath = filepath.Join(log.ContextDir, ServerConfigFile)
		})

		AfterEach(func() {
			os.Remove(serverConfigPath)
		})

		DescribeTable("server change detection",
			func(setupFunc func(), expectedChanged bool) {
				if setupFunc != nil {
					setupFunc()
				}

				ctx = &AppContext{
					ArgoServer: "current-server:443",
				}

				changed := ctx.hasServerChanged()
				Expect(changed).To(Equal(expectedChanged))
			},
			Entry("no existing config (first run)",
				func() {
					os.Remove(serverConfigPath)
				},
				false,
			),
			Entry("same server",
				func() {
					prevCtx := &AppContext{ArgoServer: "current-server:443"}
					prevCtx.saveServerConfig()
				},
				false,
			),
			Entry("different server",
				func() {
					prevCtx := &AppContext{ArgoServer: "old-server:443"}
					prevCtx.saveServerConfig()
				},
				true,
			),
			Entry("corrupted config file",
				func() {
					os.WriteFile(serverConfigPath, []byte("invalid json{{{"), 0644)
				},
				false,
			),
		)
	})

	Describe("Constants", func() {
		DescribeTable("should have correct constant values",
			func(actual, expected interface{}) {
				Expect(actual).To(Equal(expected))
			},
			Entry("ServerConfigFile", ServerConfigFile, "server_config.json"),
			Entry("ApplicationCacheFile", ApplicationCacheFile, "application_cache.json"),
			Entry("ApplicationCacheTTL", ApplicationCacheTTL, 60*time.Minute),
		)
	})
})
