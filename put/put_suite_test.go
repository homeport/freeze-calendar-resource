package put_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPut(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Put Suite")
}
