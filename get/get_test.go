package get_test

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/homeport/freeze-calendar-resource/concourse"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var check string
var err error
var tmpDir string

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
		var config io.Reader

		BeforeEach(func() {
			tmpDir = GinkgoT().TempDir()
			config = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				},
				"version": { "sha": "56dd3927d2582a332cacd5c282629293cd9a8870" }
			}`)
		})

		JustBeforeEach(func() {
			command := exec.Command(check, "get", tmpDir)
			command.Stdin = config
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

		Context("response on stdout", func() {
			var response concourse.Response

			JustBeforeEach(func() {
				err = json.Unmarshal(session.Out.Contents(), &response)
			})

			It("is valid JSON", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("has a SHA field with the expected value", func() {
				Expect(response.Version).To(HaveField("SHA", Equal("56dd3927d2582a332cacd5c282629293cd9a8870")))
			})
		})

		Context("calendar file", func() {
			var content []byte

			JustBeforeEach(func() {
				content, err = os.ReadFile(filepath.Join(tmpDir, "examples/freeze-calendar.yaml"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("has some bytes", func() {
				Expect(content).ToNot(BeEmpty())
			})
		})
	})
})
