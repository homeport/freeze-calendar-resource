package freeze_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFreeze(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Freeze Suite")
}
