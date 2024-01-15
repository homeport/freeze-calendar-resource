package get

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	timeMachine "github.com/benbjohnson/clock"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/freeze"
	"github.com/homeport/freeze-calendar-resource/lgr"
	"github.com/homeport/freeze-calendar-resource/resource"
)

type Request struct {
	resource.Request
	Version resource.Version `json:"version" validate:"required"`
	Params  resource.Params  `json:"params"`
}

type Response struct {
	Version  resource.Version         `json:"version"`
	Metadata []resource.NameValuePair `json:"metadata,omitempty"`
}

type ContextKey string

const ContextKeyClock = ContextKey("clock")
const minimumRetryInterval = 10 * time.Second

func Get(ctx context.Context, req io.Reader, resp, w io.Writer, destination string) error {
	var request Request
	err := json.NewDecoder(req).Decode(&request)

	if err != nil {
		return fmt.Errorf("unable to decode request: %w", err)
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

	if err != nil {
		return fmt.Errorf("unable to build validator: %w", err)
	}

	auth, err := request.Source.Auth()

	if err != nil {
		return fmt.Errorf("unable to build authenticator: %w", err)
	}

	logLevel := lgr.InfoLevel

	if request.Params.Verbose {
		logLevel = lgr.DebugLevel
	}

	logger := lgr.Logger{
		Level:  logLevel,
		Writer: w,
	}

	var branch = request.Source.Branch

	if branch == "" {
		branch = "main"
		logger.Debug("No branch given; falling back to %s", branch)
	}

	repo, err := git.PlainCloneContext(ctx, destination, false, &git.CloneOptions{
		URL:           request.Source.URI,
		ReferenceName: plumbing.ReferenceName(branch),
		SingleBranch:  true,
		Auth:          auth,
		Progress:      logger,
	})

	if err != nil {
		return fmt.Errorf("unable to clone: %w", err)
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return fmt.Errorf("unable to get worktree: %w", err)
	}

	head, err := repo.Head()

	if err != nil {
		return fmt.Errorf("unable to determine head: %w", err)
	}

	// Only in fuse mode we want the specific SHA that was discovered by check.
	// In gate mode we want to check out the _latest_ version of the branch,
	// which has already been provided by the initial clone.
	if request.Params.Mode == resource.Fuse {
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(request.Version.SHA),
		})

		if err != nil {
			return fmt.Errorf("unable to checkout %s: %w", request.Version.SHA, err)
		}
	}

	var totalNumberOfFreezeWindows int
	var numberOfActiveFreezeWindows int

	logger.Info("Using freeze calendar from %s at %s", request.Source.Path, head.Hash())
	var windowsPrinted []*plumbing.Reference

	for {
		calendarFile, err := worktree.Filesystem.Open(request.Source.Path)

		if err != nil {
			return fmt.Errorf("unable to read calendar file from path %s: %w", request.Source.Path, err)
		}

		calendar, err := freeze.LoadCalendar(calendarFile)

		if err != nil {
			return fmt.Errorf("unable to load calendar: %w", err)
		}

		var now time.Time

		if value := ctx.Value(ContextKeyClock); value != nil {
			now = value.(timeMachine.Clock).Now().UTC()
		} else {
			now = time.Now().UTC()
		}

		nowWithRunway := now.Add(request.Params.Runway.Duration)

		var activeFreezeWindows []freeze.Window

		for _, window := range calendar.Windows {
			if window.Start.After(nowWithRunway) {
				logger.Debug("Skipping window '%s' as its start %s is in the future (after %s + %s runway)", window.Name, window.Start.UTC(), now.UTC(), request.Params.Runway.Duration)
				continue
			}

			if window.End.Before(nowWithRunway) {
				logger.Debug("Skipping window '%s' as its end %s is in the past (before %s + %s runway)", window.Name, window.End.UTC(), now.UTC(), request.Params.Runway.Duration)
				continue
			}

			// Now we know we are within a freeze window.
			// Let's check if the scope matches. No scope for a window or the request means all windows are considered matching, as long as the dates match.
			if len(request.Params.Scope) == 0 {
				activeFreezeWindows = append(activeFreezeWindows, window)
			} else {
				for _, rs := range request.Params.Scope {
					if len(window.Scope) == 0 {
						logger.Debug("Adding window '%s' as it is not restricted to any scopes", window)
						activeFreezeWindows = append(activeFreezeWindows, window)
					} else {
						for _, ws := range window.Scope {
							if rs == ws {
								activeFreezeWindows = append(activeFreezeWindows, window)
							} else {
								logger.Debug("Skipping window '%s' as its scope %s does not match the configured scope %s", window, ws, rs)
							}
						}
					}
				}
			}
		}

		totalNumberOfFreezeWindows = len(calendar.Windows)
		numberOfActiveFreezeWindows = len(activeFreezeWindows)

		if len(activeFreezeWindows) == 0 {
			logger.Info("No active freeze windows")
			break
		} else {
			switch request.Params.Mode {
			default:
				return fmt.Errorf("unknown mode %s", request.Params.Mode)
			case resource.Fuse:
				return fmt.Errorf(
					"fuse has blown because the following freeze windows are currently active for the configured scope %s:\n%s",
					strings.Join(request.Params.Scope, ", "),
					strings.Join(mapFunc(activeFreezeWindows, func(w freeze.Window) string { return w.String() }), "\n"),
				)
			case resource.Gate:
				if !slices.Contains(windowsPrinted, head) {
					logger.Info("At %s, %d freeze windows are currently active for the configured scope %s: %s",
						head.Hash(),
						len(activeFreezeWindows),
						strings.Join(request.Params.Scope, ", "),
						strings.Join(mapFunc(activeFreezeWindows, func(w freeze.Window) string { return w.String() }), "\n"),
					)

					windowsPrinted = append(windowsPrinted, head)
				}

				newHead, err := pullAndReset(ctx, repo, branch, auth, logger)

				if err != nil {
					return err
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if newHead.Hash() != head.Hash() {
						logger.Info("Head has moved from %s to %s", head.Hash(), newHead.Hash())
						head = newHead
					} else {
						logger.Write([]byte("."))

						if request.Params.RetryInterval.Duration < minimumRetryInterval {
							time.Sleep(minimumRetryInterval)
						} else {
							time.Sleep(request.Params.RetryInterval.Duration)
						}
					}
				}
			}
		}
	}

	// re-read the _actual_ head, regardless of any branch switches made before
	head, err = repo.Head()

	if err != nil {
		return fmt.Errorf("unable to determine actual HEAD: %w", err)
	}

	response := Response{
		Version: resource.Version{SHA: head.Hash().String()},
		Metadata: []resource.NameValuePair{
			{Name: "total number of freeze windows", Value: fmt.Sprintf("%d", totalNumberOfFreezeWindows)},
			{Name: "number of active freeze windows", Value: fmt.Sprintf("%d", numberOfActiveFreezeWindows)},
		},
	}

	return json.NewEncoder(resp).Encode(response)
}

// https://stackoverflow.com/a/71624929
func mapFunc[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))

	for i := range ts {
		us[i] = f(ts[i])
	}

	return us
}

func pullAndReset(ctx context.Context, repo *git.Repository, branch string, auth transport.AuthMethod, logger lgr.Logger) (*plumbing.Reference, error) {
	err := repo.FetchContext(ctx, &git.FetchOptions{
		Auth:     auth,
		Progress: logger,
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return repo.Head() // already at the latest commit
		} else {
			return nil, fmt.Errorf("fetch failed: %w", err)
		}
	}

	remoteHead, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)

	if err != nil {
		return nil, fmt.Errorf("unable to resolve reference to branch %s: %w", branch, err)
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return nil, err
	}

	err = worktree.Reset(&git.ResetOptions{
		Commit: remoteHead.Hash(),
		Mode:   git.HardReset,
	})

	if err != nil {
		return nil, fmt.Errorf("resetting the workspace failed: %w", err)
	}

	return repo.Head()
}
