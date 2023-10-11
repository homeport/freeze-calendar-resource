package freeze_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/homeport/freeze-calendar-resource/freeze"
)

var _ = Describe("Calendar", func() {
	var (
		err      error
		calendar *freeze.Calendar
		content  string
	)

	JustBeforeEach(func() {
		calendar, err = freeze.LoadCalendar(strings.NewReader(content))
	})

	Context("empty document", func() {
		BeforeEach(func() {
			content = "---"
		})

		It("is acceptable", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("has no windows", func() {
			Expect(calendar.Windows).To(BeEmpty())
		})
	})

	Context("some windows", func() {
		BeforeEach(func() {
			content = `
freeze_calendar:
  - name: Holiday Season
    starts_at: 2022-12-01T06:00:00Z
    ends_at: 2022-12-27T06:00:00Z
    scope:
      - eu-de
      - us-east
      - ap-southeast
`
		})

		It("works", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("has some windows", func() {
			Expect(calendar.Windows).ToNot(BeEmpty())
		})

		Context("first window", func() {
			var window0 *freeze.Window

			JustBeforeEach(func() {
				window0 = &calendar.Windows[0]
			})

			It("exists", func() {
				Expect(window0).ToNot(BeNil())
			})

			It("has the expected name", func() {
				Expect(window0.Name).To(Equal("Holiday Season"))
			})

			It("has the expected start date", func() {
				Expect(window0.Start).To(Equal(time.Unix(1669874400, 0).UTC()))
			})

			It("has the expected end date", func() {
				Expect(window0.End).To(Equal(time.Unix(1672120800, 0).UTC()))
			})

			It("has the expected scope", func() {
				Expect(window0.Scope).To(HaveExactElements([]string{"eu-de", "us-east", "ap-southeast"}))
			})
		})
	})

	Context("window without name", func() {
		BeforeEach(func() {
			content = `
freeze_calendar:
  - name:
    starts_at: 2022-12-01T06:00:00Z
    ends_at: 2022-12-27T06:00:00Z
`
		})

		It("fails", func() {
			Expect(err).To(HaveOccurred())
		})

		It("has the expected error message", func() {
			Expect(err).To(MatchError(ContainSubstring("validation for 'Name' failed")))
		})
	})

	Context("ends_at is before starts_at", func() {
		BeforeEach(func() {
			content = `
freeze_calendar:
  - name: Wrong order
    starts_at: 2022-12-27T06:00:00Z
    ends_at: 2022-12-01T06:00:00Z
`
		})

		It("fails", func() {
			Expect(err).To(HaveOccurred())
		})

		It("has the expected error message", func() {
			Expect(err).To(MatchError(ContainSubstring("validation for 'End' failed")))
		})
	})

	Context("empty scope", func() {
		BeforeEach(func() {
			content = `freeze_calendar:
  - name: Holiday Season
    starts_at: 2022-12-01T06:00:00Z
    ends_at: 2022-12-27T06:00:00Z
    scope:
`
		})

		It("is acceptable", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
