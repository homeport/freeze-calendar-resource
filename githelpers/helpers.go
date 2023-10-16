package githelpers

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// https://github.com/go-git/go-git/issues/279#issuecomment-816714124
func CheckoutBranch(repo *git.Repository, name string) error {
	localTrackingBranchRef := plumbing.NewBranchReferenceName(name)

	err := repo.CreateBranch(&config.Branch{
		Name:   name,
		Remote: "origin",
		Merge:  localTrackingBranchRef,
	})

	if err != nil {
		return fmt.Errorf("unable to create local branch %s: %w", name, err)
	}

	newReference := plumbing.NewSymbolicReference(localTrackingBranchRef, plumbing.NewRemoteReferenceName("origin", name))
	err = repo.Storer.SetReference(newReference)

	if err != nil {
		return fmt.Errorf("unable to checkout local tracking branch %s: %w", newReference, err)
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return err
	}

	worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(localTrackingBranchRef.String()),
	})

	return nil
}
