package get_test

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	timeMachine "github.com/benbjohnson/clock"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/homeport/freeze-calendar-resource/get"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get in gate mode", func() {
	var (
		err            error
		req            io.Reader
		resp           strings.Builder
		log            strings.Builder
		repo           *git.Repository
		origin         string
		initialHead    plumbing.Hash
		destinationDir string
		clock          *timeMachine.Mock
		now            time.Time
		retryInterval  time.Duration
	)

	BeforeEach(func() {
		tmpDir := GinkgoT().TempDir()
		origin = path.Join(tmpDir, "remote")
		destinationDir = path.Join(tmpDir, "resource-destination-directory")
		resp = strings.Builder{}
		log = strings.Builder{}
		clock = timeMachine.NewMock()
		now = time.Unix(1691780400, 0)   // 2023-08-11T19:00:00Z
		retryInterval = 10 * time.Second // must not be below enforced minimum

		repo, err = git.PlainInitWithOptions(origin, &git.PlainInitOptions{
			InitOptions: git.InitOptions{DefaultBranch: plumbing.Main},
			Bare:        false,
		})
		Expect(err).ShouldNot(HaveOccurred())

		initialHead, err = addAndCommit(repo, "calendar.yaml", []byte(`
freeze_calendar:
  - name: Unit Test
    starts_at: 2023-07-20T09:00:00Z
    ends_at: 2023-08-20T11:00:00Z
`), "Create freeze calendar")
		Expect(err).ShouldNot(HaveOccurred())
	})

	JustBeforeEach(func(sCtx SpecContext) {
		clock.Set(now)
		ctx := context.WithValue(sCtx, get.ContextKeyClock, clock)
		ctx, cancel := context.WithTimeout(ctx, 5*retryInterval) // in a pipeline, this is taken care of by a timeout parameter on the step
		defer cancel()

		err = get.Get(ctx, req, &resp, &log, destinationDir)
	})

	Context("gate mode", func() {
		Context("request without scope", func() {
			BeforeEach(func() {
				req = strings.NewReader(fmt.Sprintf(`{
					"source": {
						"uri": "%s",
						"path": "calendar.yaml"
					},
					"version": { "sha": "%s" },
					"params": {
						"mode": "gate",
						"retry_interval": "%s",
						"verbose": true
					}
				}`, origin, initialHead, retryInterval))
			})

			It("fails", func() {
				Eventually(err).Should(HaveOccurred())
			})

			// If neither the clock advances nor the calendar gets updated, we will fail eventually
			It("has an error message indicating timeout", func() {
				Expect(err).To(MatchError(ContainSubstring("context deadline exceeded")))
			})

			Context("time has gone beyond freeze window", func() {
				BeforeEach(func() {
					now = time.Unix(1692615900, 0) // 2023-08-21T11:05:00Z
				})

				It("succeeds", func() {
					Eventually(err).ShouldNot(HaveOccurred())
				})
			})

			Context("freeze window has been shortened", func() {
				BeforeEach(func() {
					go func() { // make sure we tried unsuccessfully
						time.Sleep(3 * retryInterval)
						_, err = addAndCommit(repo, "calendar.yaml", []byte(`
freeze_calendar:
  - name: Unit Test
    starts_at: 2023-07-20T09:00:00Z
    ends_at: 2023-08-10T11:00:00Z
`), "Shorten freeze window")
						Expect(err).ShouldNot(HaveOccurred())
					}()
				})

				It("succeeds", func() {
					Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("repo: %s\nlog:\n%s", destinationDir, log.String()))
				})
			})
		})
	})
})

func addAndCommit(r *git.Repository, fileName string, content []byte, msg string) (plumbing.Hash, error) {
	w, err := r.Worktree()

	if err != nil {
		return plumbing.ZeroHash, err
	}

	f, err := w.Filesystem.Create(fileName)

	if err != nil {
		return plumbing.ZeroHash, err
	}

	_, err = f.Write(content)

	if err != nil {
		return plumbing.ZeroHash, err
	}

	_, err = w.Add(fileName)

	if err != nil {
		return plumbing.ZeroHash, err
	}

	return w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Testbild Tester",
			Email: "testbild.tester@example.org",
			When:  time.Now(),
		},
	})
}
