package check_test

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/homeport/freeze-calendar-resource/check"
	"github.com/homeport/freeze-calendar-resource/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check", func() {
	var (
		err  error
		req  io.Reader
		resp strings.Builder
		log  strings.Builder
	)

	BeforeEach(func() {
		resp = strings.Builder{}
		log = strings.Builder{}
	})

	JustBeforeEach(func(ctx SpecContext) {
		err = check.Check(ctx, req, &resp, &log)
	})

	Context("first request", func() { // Version not present
		BeforeEach(func() {
			req = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				}
			}`)
		})

		It("executes successfully", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("response", func() {
			var response check.Response

			JustBeforeEach(func() {
				err = json.NewDecoder(strings.NewReader(resp.String())).Decode(&response)
			})

			It("is valid JSON", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("have at least one version", func() {
				Expect(response).ToNot(BeEmpty())
			})

			Context("latest version", func() {
				var version resource.Version

				JustBeforeEach(func() {
					version = response[0]
				})

				It("produces valid JSON with a SHA field", func() {
					Expect(version.SHA).NotTo(BeEmpty())
				})
			})
		})
	})

	Context("subsequent requests", func() {
		BeforeEach(func() {
			req = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				},
				"version": { "sha": "56dd3927d2582a332cacd5c282629293cd9a8870" }
			}`)
		})

		It("executes successfully", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("response", func() {
			var response check.Response

			JustBeforeEach(func() {
				err = json.NewDecoder(strings.NewReader(resp.String())).Decode(&response)
			})

			It("produces valid JSON", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("have at least one version", func() {
				Expect(response).ToNot(BeEmpty())
			})

			Context("requested version", func() {
				var version resource.Version

				JustBeforeEach(func() {
					version = response[0]
				})

				It("produces valid JSON with a SHA field", func() {
					Expect(version.SHA).NotTo(BeEmpty())
				})

				It("produces the expected SHA", func() {
					Expect(version.SHA).To(Equal("56dd3927d2582a332cacd5c282629293cd9a8870"))
				})
			})
		})
	})
})
