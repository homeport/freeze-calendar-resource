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
		Context("requesting an old version", func() {
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

				It("have at least two versions", func() {
					Expect(len(response)).To(BeNumerically(">", 1))
				})

				Context("oldest version", func() {
					var oldestVersion resource.Version

					JustBeforeEach(func() {
						oldestVersion = response[0]
					})

					It("is the requested one", func() {
						Expect(oldestVersion.SHA).To(Equal("56dd3927d2582a332cacd5c282629293cd9a8870"))
					})
				})

				Context("latest version", func() {
					var latestVersion resource.Version

					JustBeforeEach(func() {
						latestVersion = response[len(response)-1]
					})

					It("has the expected SHA", func() {
						Expect(latestVersion.SHA).To(Equal("6d78528138da1a6f536601d30a3967a4004b71b7"))
					})
				})
			})
		})

		Context("requesting a version later than the earliest", func() {
			BeforeEach(func() {
				req = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				},
				"version": { "sha": "6d78528138da1a6f536601d30a3967a4004b71b7" }
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

				Context("oldest version", func() {
					var oldestVersion resource.Version

					JustBeforeEach(func() {
						oldestVersion = response[0]
					})

					It("is the requested one", func() {
						Expect(oldestVersion.SHA).To(Equal("6d78528138da1a6f536601d30a3967a4004b71b7"))
					})
				})

				Context("latest version", func() {
					var latestVersion resource.Version

					JustBeforeEach(func() {
						latestVersion = response[len(response)-1]
					})

					// This will break when we ever update examples/freeze-calendar.yaml
					It("is the expected one", func() {
						Expect(latestVersion.SHA).To(Equal("6d78528138da1a6f536601d30a3967a4004b71b7"))
					})
				})
			})
		})
	})
})
