package get_test

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/homeport/freeze-calendar-resource/resource"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scope", func() {
	var (
		err     error
		request io.Reader
		params  *resource.Params
	)

	BeforeEach(func() {
		params = nil
	})

	JustBeforeEach(func() {
		err = json.NewDecoder(request).Decode(&params)
	})

	Context("some scope members", func() {
		BeforeEach(func() {
			request = strings.NewReader(`{"scope": ["a", "b", "c"]}`)
		})

		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("has params", func() {
			Expect(params).ToNot(BeNil())
		})

		It("has the expected members", func() {
			Expect(params.Scope).To(ContainElements([]string{"a", "b", "c"}))
		})
	})

	Context("no scope present", func() {
		BeforeEach(func() {
			request = strings.NewReader(`{}`)
		})

		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("has params", func() {
			Expect(params).ToNot(BeNil())
		})

		It("has no members", func() {
			Expect(params.Scope).To(BeEmpty())
		})
	})

	Context("scope empty", func() {
		BeforeEach(func() {
			request = strings.NewReader(`{"scope": []}`)
		})

		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("has params", func() {
			Expect(params).ToNot(BeNil())
		})

		It("has no members", func() {
			Expect(params.Scope).To(BeEmpty())
		})
	})
})
