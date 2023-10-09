package get

import (
	"encoding/json"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/homeport/freeze-calendar-resource/concourse"
	"github.com/spf13/cobra"
)

// TODO Metadata
func Run(cmd *cobra.Command, args []string) error {
	request, err := concourse.LoadRequest(cmd.InOrStdin())

	if err != nil {
		return err
	}

	err = concourse.ValidateRequest(request)

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

	response := concourse.Response{
		Version: request.Version,
	}

	json.NewEncoder(cmd.OutOrStdout()).Encode(response)

	return nil
}
