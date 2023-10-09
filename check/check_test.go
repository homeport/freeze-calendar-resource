package check_test

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/homeport/freeze-calendar-resource/resource"
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
		var request io.Reader

		BeforeEach(func() {
			request = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				}
			}`)
		})

		JustBeforeEach(func() {
			command := exec.Command(check, "check")
			command.Stdin = request
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())

			session.Wait(10 * time.Second)
		})

		It("executes successfully", func() {
			Expect(session.ExitCode()).To(Equal(0))
		})

		It("produces non-empty output on StdOut", func() {
			Expect(session.Out.Contents()).ToNot(BeEmpty())
		})

		It("produces no output on StdErr", func() {
			Expect(string(session.Err.Contents())).To(BeEmpty())
		})

		Context("On StdOut", func() {
			var version resource.Version

			JustBeforeEach(func() {
				err = json.NewDecoder(session.Out).Decode(&version)
			})

			It("produces valid JSON on StdOut", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("produces valid JSON with a SHA field on StdOut", func() {
				Expect(version.SHA).NotTo(BeEmpty())
			})

			It("produces the expected SHA", func() {
				Expect(version.SHA).To(Equal("56dd3927d2582a332cacd5c282629293cd9a8870"))
			})
		})
	})
})
