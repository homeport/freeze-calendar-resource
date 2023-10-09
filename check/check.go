package check

import (
	"encoding/json"
	"fmt"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/homeport/freeze-calendar-resource/concourse"
	"github.com/spf13/cobra"
)

// Expected on STDIN:
//
//	{
//	   "source": {
//		    "uri": "git@github.com:homeport/freeze-calendar-resource"
//		    "branch": "main"
//		    "private_key": "((vault/my-key))"
//		    "path": "examples/freeze-calendar.yaml"
//	   },
//	   "version": { "sha": "..." }
//	}
func Run(cmd *cobra.Command, args []string) error {
	request, err := concourse.LoadRequest(cmd.InOrStdin())

	if err != nil {
		return err
	}

	err = concourse.ValidateRequest(request)

	if err != nil {
		return err
	}

	var worktree billy.Filesystem // leaving this as nil so that we get a bare repo

	repo, err := git.Clone(memory.NewStorage(), worktree, &git.CloneOptions{
		URL: request.Source.URI,
	})

	if err != nil {
		return fmt.Errorf("unable to clone: %w", err)
	}

	cIter, err := repo.Log(&git.LogOptions{PathFilter: func(s string) bool {
		return s == request.Source.Path
	}})

	if err != nil {
		return err
	}

	commit, err := cIter.Next()

	if err != nil {
		return err
	}

	response := concourse.Version{
		SHA: commit.Hash.String(),
	}

	json.NewEncoder(cmd.OutOrStdout()).Encode(response)

	return nil
}
