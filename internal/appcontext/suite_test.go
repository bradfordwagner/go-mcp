package appcontext

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"template_cli/internal/log"
)

func TestAppContext(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppContext Suite")
}

var _ = BeforeSuite(func() {
	// Initialize logger for tests
	err := log.Init()
	if err != nil {
		// Log initialization may fail in test environment, which is acceptable
		GinkgoWriter.Printf("Warning: Failed to initialize logger: %v\n", err)
	}
})

