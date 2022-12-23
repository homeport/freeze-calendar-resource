package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

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
	URI    string `json:"uri"`
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
		log.Fatal(err)
	}
}
func mainE() error {
	config, err := LoadConfig(os.Stdin)

	if err != nil {
		return err
	}

	// TODO Do not check out the worktree
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: config.Source.URI,
	})

	if err != nil {
		return err
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

func LoadConfig(r io.Reader) (Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	return config, err
}
