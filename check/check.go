package check

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/githelpers"
	"github.com/homeport/freeze-calendar-resource/resource"
	"golang.org/x/exp/slices"
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
		return fmt.Errorf("unable to decode request: %w", err)
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

	if err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	auth, err := request.Source.Auth()

	if err != nil {
		return fmt.Errorf("unable to build authenticator: %w", err)
	}

	fs := memfs.New()

	repo, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:      request.Source.URI,
		Auth:     auth,
		Progress: log,
	})

	if err != nil {
		return fmt.Errorf("unable to clone: %w", err)
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
		fmt.Fprintf(log, "checking out branch %s\n", request.Source.Branch)

		branchName := plumbing.NewRemoteReferenceName("origin", request.Source.Branch)

		_, err := repo.Branch(string(branchName))

		if err != nil {
			return fmt.Errorf("%s %w", branchName, err) // branch does not exist
		}

		err = githelpers.CheckoutBranch(repo, branchName.Short())

		if err != nil {
			return fmt.Errorf("unable to checkout %s: %w", branchName, err)
		}
	}

	cIter, err := repo.Log(&git.LogOptions{
		PathFilter: func(s string) bool {
			return s == request.Source.Path
		},
		Order: git.LogOrderCommitterTime,
	})

	if err != nil {
		return fmt.Errorf("could not log the history: %w", err)
	}

	// "The list may be empty, if there are no versions available at the source."
	// TODO When would that happen? If the repo or branch doesn't exist?

	var response []resource.Version

	err = cIter.ForEach(func(commit *object.Commit) error {
		sha := commit.Hash.String()
		response = append(response, resource.Version{SHA: sha})
		return nil
	})

	if err != nil {
		return fmt.Errorf("could not iterate over commits: %w", err)
	}

	// "... must print the array of new versions, in chronological order (oldest first)"
	// from https://concourse-ci.org/implementing-resource-types.html#resource-check
	slices.Reverse(response)

	// If a version is provided in the request, return only versions newer than the requested one
	if request.Version.SHA != "" {
		i, found := slices.BinarySearchFunc(response, resource.Version{SHA: request.Version.SHA}, func(a, b resource.Version) int {
			return cmp.Compare(a.SHA, b.SHA)
		})

		if found {
			response = response[i:]
		} else {
			// "If your resource is unable to determine which versions are newer than the given version (e.g. if it's a git commit that was push -fed over), then the current version of your resource should be returned (i.e. the new HEAD)."
			response = []resource.Version{response[len(response)-1]}
		}
	}

	return json.NewEncoder(resp).Encode(response)
}
