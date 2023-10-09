package concourse

import (
	"encoding/json"
	"fmt"
	"io"
)

type Request struct {
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

type Response struct {
	Version Version `json:"version"`
}

// TODO Perhaps replace with https://github.com/go-playground/validator
func ValidateRequest(c Request) error {
	if c.Source.URI == "" {
		return fmt.Errorf("source.uri must not be empty")
	}

	if c.Source.Path == "" {
		return fmt.Errorf("source.path must not be empty")
	}

	return nil
}

func LoadRequest(r io.Reader) (Request, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Request{}, err
	}

	var config Request
	err = json.Unmarshal(data, &config)
	return config, err
}
