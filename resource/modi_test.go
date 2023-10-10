package resource_test

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/homeport/freeze-calendar-resource/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Modi", func() {
	var (
		err     error
		request io.Reader
		params  resource.Params
	)

	JustBeforeEach(func() {
		err = json.NewDecoder(request).Decode(&params)
	})

	Context("fuse mode", func() {
		BeforeEach(func() {
			request = strings.NewReader(`{"mode": "fuse"}`)
		})

		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("has params", func() {
			Expect(params).ToNot(BeNil())
		})

		It("has the expected mode", func() {
			Expect(params.Mode).To(Equal(resource.Fuse))
		})
	})

	Context("gate mode", func() {
		BeforeEach(func() {
			request = strings.NewReader(`{"mode": "gate"}`)
		})

		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("has params", func() {
			Expect(params).ToNot(BeNil())
		})

		It("has the expected mode", func() {
			Expect(params.Mode).To(Equal(resource.Gate))
		})
	})

	Context("unknown mode", func() {
		BeforeEach(func() {
			request = strings.NewReader(`{"mode": "foobar"}`)
		})

		It("rejects", func() {
			Expect(err).To(HaveOccurred())
		})

		It("has a useful error message", func() {
			Expect(err).To(MatchError(ContainSubstring("is not a valid mode")))
		})
	})
})
