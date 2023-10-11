package resource_test

import (
	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Source Validation", func() {
	var (
		err    error
		source resource.Source
	)

	JustBeforeEach(func() {
		err = validator.New(validator.WithRequiredStructEnabled()).Struct(source)
	})

	Context("empty source", func() {
		BeforeEach(func() {
			source = resource.Source{}
		})

		It("fails", func() {
			Expect(err).To(HaveOccurred())
		})

		It("has the expected error", func() {
			Expect(err).To(MatchError(ContainSubstring("failed")))
		})
	})

	Context("MVP", func() {
		BeforeEach(func() {
			source = resource.Source{
				URI:  "git@github.com:homeport/freeze-calendar-resource",
				Path: "/foo/bar/something.else",
			}
		})

		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("missing Path", func() {
		BeforeEach(func() {
			source = resource.Source{
				URI: "git@github.com:homeport/freeze-calendar-resource",
			}
		})

		It("fails", func() {
			Expect(err).To(HaveOccurred())
		})

		It("has the expected error", func() {
			Expect(err).To(MatchError(ContainSubstring("validation for 'Path' failed")))
		})
	})

	Context("missing URI", func() {
		BeforeEach(func() {
			source = resource.Source{
				Path: "/foo/bar/something.else",
			}
		})

		It("fails", func() {
			Expect(err).To(HaveOccurred())
		})

		It("has the expected error", func() {
			Expect(err).To(MatchError(ContainSubstring("validation for 'URI' failed")))
		})
	})
})
