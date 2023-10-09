package check

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
)

type Config struct {
	Version Version `json:"version,omitempty"`
	Source  Source  `json:"source"`
}

type Version struct {
	SHA string `json:"sha"`
}

type Source struct {
	URI    string `json:"uri"` // the git resource calls it uri, so we do it, too
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

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
	config, err := loadConfig(cmd.InOrStdin())

	if err != nil {
		return err
	}

	err = validateConfig(config)

	if err != nil {
		return err
	}

	var worktree billy.Filesystem // leaving this as nil so that we get a bare repo

	repo, err := git.Clone(memory.NewStorage(), worktree, &git.CloneOptions{
		URL: config.Source.URI,
	})

	if err != nil {
		return fmt.Errorf("unable to clone: %w", err)
	}

	cIter, err := repo.Log(&git.LogOptions{PathFilter: func(s string) bool {
		return s == config.Source.Path
	}})

	if err != nil {
		return err
	}

	commit, err := cIter.Next()

	if err != nil {
		return err
	}

	fmt.Printf(`{"sha": "%s"}`, commit.Hash.String())

	return nil
}

// TODO Perhaps replace with https://github.com/go-playground/validator
func validateConfig(c Config) error {
	if c.Source.URI == "" {
		return fmt.Errorf("source.uri must not be empty")
	}

	if c.Source.Path == "" {
		return fmt.Errorf("source.path must not be empty")
	}

	return nil
}

func loadConfig(r io.Reader) (Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	return config, err
}
