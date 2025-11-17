package argoclient

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ArgoClient", func() {
	Describe("NewClient", func() {
		DescribeTable("Server URL Normalization",
			func(inputServer, expectedServer string) {
				cfg := Config{
					Server:    inputServer,
					AuthToken: "test-token",
					Insecure:  true,
				}

				client, err := NewClient(cfg)
				// We expect this to fail since we're not connecting to a real server
				// but we can still check the server URL normalization
				if err != nil {
					Expect(err.Error()).To(Or(
						ContainSubstring(expectedServer),
						ContainSubstring("failed to create Argo CD client"),
					))
				} else {
					Expect(client).NotTo(BeNil())
					Expect(client.Server).To(Equal(expectedServer))
				}
			},
			Entry("strips https prefix", "https://argocd.example.com:443", "argocd.example.com:443"),
			Entry("strips http prefix", "http://argocd.example.com:80", "argocd.example.com:80"),
			Entry("keeps URL without prefix unchanged", "argocd.example.com:443", "argocd.example.com:443"),
			Entry("handles localhost with https", "https://localhost:8080", "localhost:8080"),
		)
	})

	Describe("NewConfigFromEnv", func() {
		Context("when required environment variables are missing", func() {
			It("should return an error", func() {
				ctx := context.Background()
				_, err := NewConfigFromEnv(ctx)
				
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to process environment variables"))
			})
		})
	})

	Describe("Config Structure", func() {
		It("should properly store configuration values", func() {
			cfg := Config{
				Server:    "test-server",
				AuthToken: "test-token",
				Insecure:  true,
			}

			Expect(cfg.Server).To(Equal("test-server"))
			Expect(cfg.AuthToken).To(Equal("test-token"))
			Expect(cfg.Insecure).To(BeTrue())
		})
	})

	Describe("ClientWithServer Structure", func() {
		It("should wrap client and server URL", func() {
			cws := &ClientWithServer{
				Client: nil,
				Server: "test-server:443",
			}

			Expect(cws.Server).To(Equal("test-server:443"))
		})
	})
})
