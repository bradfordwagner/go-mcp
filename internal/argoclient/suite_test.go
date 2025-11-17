package argoclient

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestArgoClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ArgoClient Suite")
}

