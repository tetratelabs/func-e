package push_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPush(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Suite")
}
