package check

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/resource"
)

type Request struct {
	resource.Request
	Version resource.Version `json:"version"`
}

type Response []resource.Version

// Request:
//
//	{
//	   "source": {
//		    "uri": "git@github.com:homeport/freeze-calendar-resource"
//		    "branch": "main"
//		    "private_key": "((vault/my-key))"
//		    "path": "examples/freeze-calendar.yaml"
//	   },
//	   "version": { "sha": "..." } // may be present or not
//	}
//
// Response:
//
// [{ "version": { "sha": "..." } }]
func Check(ctx context.Context, req io.Reader, resp, log io.Writer) error {
	var request Request
	err := json.NewDecoder(req).Decode(&request)

	if err != nil {
		return err
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

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

	// TODO
	// The list may be empty, if there are no versions available at the source.
	// If the given version is already the latest, an array with that version as the sole entry should be listed.
	// If your resource is unable to determine which versions are newer than the given version (e.g. if it's a git commit that was push -fed over), then the current version of your resource should be returned (i.e. the new HEAD).

	response := []resource.Version{{SHA: commit.Hash.String()}}

	return json.NewEncoder(resp).Encode(response)
}
