package put_test

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/homeport/freeze-calendar-resource/put"
)

var _ = Describe("Put", func() {
	var (
		err    error
		req    io.Reader
		resp   strings.Builder
		log    strings.Builder
		tmpDir string
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		req = strings.NewReader(`{
		  "source": {
        "uri": "https://github.com/homeport/freeze-calendar-resource",
        "path": "examples/freeze-calendar.yaml"
      },
      "version": { "sha": "56dd3927d2582a332cacd5c282629293cd9a8870" },
      "params": { "mode": "fuse", "scope": ["eu-de"] }
    }`)

		resp = strings.Builder{}
		log = strings.Builder{}
	})

	JustBeforeEach(func(ctx SpecContext) {
		err = put.Put(ctx, req, &resp, &log, tmpDir)
	})

	It("executes successfully", func() {
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("response", func() {
		var response put.Response

		JustBeforeEach(func() {
			err = json.NewDecoder(strings.NewReader(resp.String())).Decode(&response)
		})

		It("is valid JSON", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("source directory", func() {
		var entries []os.DirEntry

		JustBeforeEach(func() {
			entries, err = os.ReadDir(tmpDir)
			Expect(err).ToNot(HaveOccurred())
		})

		It("is empty", func() {
			Expect(entries).To(BeEmpty())
		})
	})
})
