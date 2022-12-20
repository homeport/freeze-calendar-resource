package main_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var check string
var err error

var _ = BeforeSuite(func() {
	check, err = gexec.Build("github.com/homeport/freeze-calendar-resource")
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("Check", func() {
	It("builds successfully", func() {
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("Executing command", func() {
		var session *gexec.Session

		JustBeforeEach(func() {
			command := exec.Command(check)
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())

			session.Wait()
		})

		It("executes successfully", func() {
			Expect(session.ExitCode()).To(Equal(0))
		})

		It("produces non-empty output on StdOut", func() {
			Eventually(session.Out.Contents()).ShouldNot(BeEmpty())
		})

		It("produces no output on StdErr", func() {
			Eventually(session.Err.Contents()).Should(BeEmpty())
		})

		Context("On StdOut", func() {
			var version struct {
				SHA string `json:"sha"`
			}

			JustBeforeEach(func() {
				err = json.Unmarshal(session.Out.Contents(), &version)
			})

			It("produces valid JSON on StdOut", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("produces valid JSON with a SHA field on StdOut", func() {
				Expect(version.SHA).NotTo(BeEmpty())
			})
		})
	})
})
