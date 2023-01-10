package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
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

//	{
//	   "source": {
//		    "uri": "git@github.com:homeport/freeze-calendar-resource"
//		    "branch": "main"
//		    "private_key": "((vault/my-key))"
//		    "path": "examples/freeze-calendar.yaml"
//	   },
//	   "version": { "sha": "..." }
//	}
func main() {
	err := mainE()

	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
func mainE() error {
	config, err := loadConfig(os.Stdin)

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
