package get

import (
	"encoding/json"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/homeport/freeze-calendar-resource/freeze"
	"github.com/homeport/freeze-calendar-resource/resource"
	"github.com/spf13/cobra"
)

type Request struct {
	Version resource.Version `json:"version,omitempty"`
	Source  resource.Source  `json:"source"`
	Params  resource.Params  `json:"params"`
}

func Run(cmd *cobra.Command, args []string) error {
	var request Request
	err := json.NewDecoder(cmd.InOrStdin()).Decode(&request)

	if err != nil {
		return err
	}

	err = resource.ValidateSource(request.Source)

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
		return fmt.Errorf("unable to decode calendar: %w", err)
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
