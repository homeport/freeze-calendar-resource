package get

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	timeMachine "github.com/benbjohnson/clock"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/freeze"
	"github.com/homeport/freeze-calendar-resource/resource"
	"github.com/orsinium-labs/enum"
)

type GetRequest struct {
	resource.Request
	Params Params `json:"params"`
}

type Params struct {
	Mode  Mode     `json:"mode" validate:"required"`
	Scope []string `json:"scope"`
	Debug bool     `json:"debug"`
}

type Mode enum.Member[string]

var (
	Fuse = Mode{"fuse"}
	Gate = Mode{"gate"}
	Modi = enum.New(Fuse, Gate)
)

type ContextKey string

const ContextKeyClock = ContextKey("clock")

func (m *Mode) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))

	if err != nil {
		return err
	}

	parsed := Modi.Parse(unquoted)

	if parsed == nil {
		return fmt.Errorf("%s is not a valid mode, valid ones are %s", string(b), Modi.String())
	}

	*m = *parsed
	return nil
}

func Get(ctx context.Context, req io.Reader, resp, log io.Writer, destination string) error {
	var request GetRequest
	err := json.NewDecoder(req).Decode(&request)

	if err != nil {
		return err
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

	if err != nil {
		return err
	}

	repo, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL: request.Source.URI,
	})

	if err != nil {
		return fmt.Errorf("unable to clone: %w", err)
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return fmt.Errorf("unable to get worktree: %w", err)
	}

	if request.Version.SHA != "" {
		err = worktree.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(request.Version.SHA),
		})
	}

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
			fmt.Fprintf(log, "Skipping window '%s' as its start %s is in the future (after %s)", window.Name, window.Start.UTC(), now.UTC())
			continue
		}

		if window.End.Before(now) {
			fmt.Fprintf(log, "Skipping window '%s' as its end %s is in the past (before %s)", window.Name, window.End.UTC(), now.UTC())
			continue
		}

		// Now we know we are within a freeze window.
		// Let's check if the scope matches.
		for _, rs := range request.Params.Scope {
			for _, ws := range window.Scope {
				if rs == ws {
					activeFreezeWindows = append(activeFreezeWindows, window)
				}
			}
		}
	}

	if len(activeFreezeWindows) > 0 {
		switch request.Params.Mode {
		case Fuse:
			return fmt.Errorf(
				"fuse has blown because the following freeze windows are currently active for the scope %s: %s",
				strings.Join(request.Params.Scope, ", "),
				strings.Join(mapFunc(activeFreezeWindows, func(w freeze.Window) string { return w.String() }), ", "),
			)
		case Gate:
			return errors.New("gate mode not implemented yet")
		default:
			return fmt.Errorf("unknown mode %s", request.Params.Mode)
		}
	}

	ref, err := repo.Head()

	if err != nil {
		return fmt.Errorf("unable to determine HEAD: %w", err)
	}

	response := resource.Response{
		Version: resource.Version{SHA: ref.Hash().String()},
		Metadata: []resource.NameValuePair{
			{Name: "number of freeze windows", Value: fmt.Sprintf("%d", len(calendar.Windows))},
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
