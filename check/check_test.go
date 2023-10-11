package check_test

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/homeport/freeze-calendar-resource/check"
	"github.com/homeport/freeze-calendar-resource/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Check", func() {
	var (
		err    error
		cmd    *cobra.Command
		stdin  io.Reader
		stdout strings.Builder
		stderr strings.Builder
	)

	BeforeEach(func() {
		stdin = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				}
			}`)
		stdout = strings.Builder{}
		stderr = strings.Builder{}
		cmd = &cobra.Command{RunE: check.RunE}
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{}) // Don't let Ginkgo's arguments get in the way
	})

	JustBeforeEach(func(ctx SpecContext) {
		cmd.SetIn(stdin)
		err = cmd.ExecuteContext(ctx)
	})

	It("executes successfully", func() {
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("response on stdout", func() {
		var version resource.Version

		JustBeforeEach(func() {
			err = json.NewDecoder(strings.NewReader(stdout.String())).Decode(&version)
		})

		It("produces valid JSON on stdout", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("produces valid JSON with a SHA field on stdout", func() {
			Expect(version.SHA).NotTo(BeEmpty())
		})

		It("produces the expected SHA", func() {
			Expect(version.SHA).To(Equal("56dd3927d2582a332cacd5c282629293cd9a8870"))
		})
	})
})
