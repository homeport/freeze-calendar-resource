package get

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	timeMachine "github.com/benbjohnson/clock"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/freeze"
	"github.com/homeport/freeze-calendar-resource/githelpers"
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

func Get(ctx context.Context, req io.Reader, resp, log io.Writer, destination string) error {
	var request Request
	err := json.NewDecoder(req).Decode(&request)

	if err != nil {
		return fmt.Errorf("unable to build decoder: %w", err)
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

	if err != nil {
		return fmt.Errorf("unable to build validator: %w", err)
	}

	auth, err := request.Source.Auth()

	if err != nil {
		return fmt.Errorf("unable to build authenticator: %w", err)
	}

	repo, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL:      request.Source.URI,
		Auth:     auth,
		Progress: log,
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

	// Assuming that head.Name().IsBranch() is always true

	shortHeadName := head.Name().Short()

	if request.Source.Branch == "" {
		fmt.Fprintf(log, "using default remote branch %s\n", shortHeadName)
	} else if request.Source.Branch == shortHeadName {
		fmt.Fprintf(log, "%s already checked out\n", shortHeadName)
	} else {
		fmt.Fprintf(log, "checking branch %s\n", request.Source.Branch)
		err = githelpers.CheckoutBranch(repo, request.Source.Branch)

		if err != nil {
			return fmt.Errorf("unable to switch to branch %s: %w", request.Source.Branch, err)
		}
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(request.Version.SHA),
	})

	if err != nil {
		return fmt.Errorf("unable to checkout %s: %w", request.Version.SHA, err)
	}

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

	var activeFreezeWindows []freeze.Window

	for _, window := range calendar.Windows {
		if window.Start.After(now) {
			fmt.Fprintf(log, "Skipping window '%s' as its start %s is in the future (after %s)\n", window.Name, window.Start.UTC(), now.UTC())
			continue
		}

		if window.End.Before(now) {
			fmt.Fprintf(log, "Skipping window '%s' as its end %s is in the past (before %s)\n", window.Name, window.End.UTC(), now.UTC())
			continue
		}

		// Now we know we are within a freeze window.
		// Let's check if the scope matches. No scope in a window or the request means all windows are considered matching, as long as the dates match.
		if len(request.Params.Scope) == 0 {
			activeFreezeWindows = append(activeFreezeWindows, window)
		} else {
			for _, rs := range request.Params.Scope {
				if len(window.Scope) == 0 {
					fmt.Fprintf(log, "Adding window '%s' as it is not restricted to any scopes\n", window)
					activeFreezeWindows = append(activeFreezeWindows, window)
				} else {
					for _, ws := range window.Scope {
						if rs == ws {
							activeFreezeWindows = append(activeFreezeWindows, window)
						} else {
							fmt.Fprintf(log, "Skipping window '%s' as its scope %s does not match the configured scope %s\n", window, ws, rs)
						}
					}
				}
			}
		}
	}

	if len(activeFreezeWindows) == 0 {
		fmt.Fprintln(log, "No windows matching")
	} else {
		switch request.Params.Mode {
		case resource.Fuse:
			return fmt.Errorf(
				"fuse has blown because the following freeze windows are currently active for the configured scope %s:\n%s",
				strings.Join(request.Params.Scope, ", "),
				strings.Join(mapFunc(activeFreezeWindows, func(w freeze.Window) string { return w.String() }), "\n"),
			)
		case resource.Gate:
			return errors.New("gate mode not implemented yet")
		default:
			return fmt.Errorf("unknown mode %s", request.Params.Mode)
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
			{Name: "total number of freeze windows", Value: fmt.Sprintf("%d", len(calendar.Windows))},
			{Name: "number of active freeze windows", Value: fmt.Sprintf("%d", len(activeFreezeWindows))},
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
