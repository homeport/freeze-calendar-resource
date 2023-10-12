package get_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	timeMachine "github.com/benbjohnson/clock"

	"github.com/homeport/freeze-calendar-resource/get"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get", func() {
	var (
		err    error
		req    io.Reader
		resp   strings.Builder
		log    strings.Builder
		tmpDir string
		clock  *timeMachine.Mock
		now    time.Time
	)

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()

		resp = strings.Builder{}
		log = strings.Builder{}
		clock = timeMachine.NewMock()
	})

	JustBeforeEach(func(ctx SpecContext) {
		clock.Set(now)
		err = get.Get(
			context.WithValue(ctx, get.ContextKeyClock, clock),
			req,
			&resp,
			&log,
			tmpDir,
		)
	})

	Context("mode not specified", func() {
		BeforeEach(func() {
			req = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				},
				"params": {  },
				"version": { "sha": "56dd3927d2582a332cacd5c282629293cd9a8870" }
			}`)
		})

		It("fails", func() {
			Expect(err).To(HaveOccurred())
		})

		It("has an error message", func() {
			Expect(err).To(MatchError(ContainSubstring("validation for 'Mode' failed")))
		})
	})

	Context("fuse mode", func() {
		BeforeEach(func() {
			req = strings.NewReader(`{
				"source": {
					"uri": "https://github.com/homeport/freeze-calendar-resource",
					"path": "examples/freeze-calendar.yaml"
				},
				"version": { "sha": "56dd3927d2582a332cacd5c282629293cd9a8870" },
				"params": { "mode": "fuse", "scope": ["eu-de"] }
			}`)
		})

		It("executes successfully", func() {
			Expect(err).ShouldNot(HaveOccurred())
		})

		Context("response", func() {
			var response get.Response

			JustBeforeEach(func() {
				err = json.NewDecoder(strings.NewReader(resp.String())).Decode(&response)
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

		Context("within the freeze", func() {
			BeforeEach(func(ctx SpecContext) {
				now = time.Unix(1671690195, 0)
			})

			Context("in scope", func() {
				It("fails", func() {
					Expect(err).To(HaveOccurred())
				})

				It("has an error message", func() {
					Expect(err).To(MatchError(ContainSubstring("fuse has blown")))
				})
			})

			Context("out of scope", func() {
				BeforeEach(func() {
					req = strings.NewReader(`{
						"source": {
							"uri": "https://github.com/homeport/freeze-calendar-resource",
							"path": "examples/freeze-calendar.yaml"
						},
						"version": { "sha": "56dd3927d2582a332cacd5c282629293cd9a8870" },
						"params": { "mode": "fuse", "scope": ["eu-gb"] }
					}`)
				})

				It("succeeds", func() {
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
