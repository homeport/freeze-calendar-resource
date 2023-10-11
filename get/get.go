package get

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/freeze"
	"github.com/homeport/freeze-calendar-resource/resource"
	"github.com/orsinium-labs/enum"
	"github.com/spf13/cobra"
)

type Request struct {
	Version resource.Version `json:"version,omitempty" validate:"required"`
	Source  resource.Source  `json:"source" validate:"required"`
	Params  Params           `json:"params"`
}

type Params struct {
	Mode  Mode     `json:"mode"`
	Scope []string `json:"scope"`
	Debug bool     `json:"debug"`
}

type Mode enum.Member[string]

var (
	Fuse = Mode{"fuse"}
	Gate = Mode{"gate"}
	Modi = enum.New(Fuse, Gate)
)

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

func RunE(cmd *cobra.Command, args []string) error {
	var request Request
	err := json.NewDecoder(cmd.InOrStdin()).Decode(&request)

	if err != nil {
		return err
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

	if err != nil {
		return err
	}

	repo, err := git.PlainClone(args[0], false, &git.CloneOptions{
		URL: request.Source.URI,
	})

	if err != nil {
		return fmt.Errorf("unable to clone: %w", err)
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return fmt.Errorf("unable to get worktree: %w", err)
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

	now := time.Now() // TODO use clock for testability
	var activeFreezeWindows []freeze.Window

	for _, window := range calendar.Windows {
		if window.Start.After(now) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Skipping window '%s' as its start %s is in the future (after %s)", window.Name, window.Start.UTC(), now.UTC())
			continue
		}

		if window.End.Before(now) {
			fmt.Fprintf(cmd.ErrOrStderr(), "Skipping window '%s' as its end %s is in the past (before %s)", window.Name, window.End.UTC(), now.UTC())
			continue
		}

		// Now we know we are within a freeze window.
		activeFreezeWindows = append(activeFreezeWindows, window)

		// TODO Let's check if the scope matches.
	}

	if len(activeFreezeWindows) > 0 {
		if request.Params.Mode == Fuse {
			return fmt.Errorf("fuse has blown because the following freeze windows are currently active: %s", strings.Join(mapFunc(activeFreezeWindows, func(w freeze.Window) string {
				return w.String()
			}), ", "))
		}
	}

	response := resource.Response{
		Version: request.Version,
		Metadata: []resource.NameValuePair{
			{Name: "number of freeze windows", Value: fmt.Sprintf("%d", len(calendar.Windows))},
		},
	}

	json.NewEncoder(cmd.OutOrStdout()).Encode(response)

	return nil
}

// https://stackoverflow.com/a/71624929
func mapFunc[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))

	for i := range ts {
		us[i] = f(ts[i])
	}

	return us
}
