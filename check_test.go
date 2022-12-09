package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var err error

var _ = BeforeSuite(func() {
	_, err = gexec.Build("github.com/homeport/freeze-calendar-resource")
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("Check", func() {

	It("executes successfully", func() {
		Expect(err).ShouldNot(HaveOccurred())

		// Expect(exitCode).To(Equal(0))
	})
})
