package get_test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/homeport/freeze-calendar-resource/get"
	"github.com/homeport/freeze-calendar-resource/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Get", func() {
	var (
		err    error
		cmd    *cobra.Command
		stdin  io.Reader
		stdout strings.Builder
		stderr strings.Builder
		tmpDir string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		stdin = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				},
				"version": { "sha": "56dd3927d2582a332cacd5c282629293cd9a8870" },
				"params": { "mode": "fuse" }
			}`)

		stdout = strings.Builder{}
		stderr = strings.Builder{}

		cmd = &cobra.Command{RunE: get.RunE}
		cmd.SetArgs([]string{tmpDir})
		cmd.SetIn(stdin)
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
	})

	JustBeforeEach(func() {
		err = cmd.Execute()
	})

	It("executes successfully", func() {
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("response on stdout", func() {
		var response resource.Response

		JustBeforeEach(func() {
			err = json.NewDecoder(strings.NewReader(stdout.String())).Decode(&response)
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
