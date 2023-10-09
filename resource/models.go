package resource

import (
	"fmt"
)

type Version struct {
	SHA string `json:"sha"`
}

type Source struct {
	URI    string `json:"uri"` // the git resource calls it uri, so we do it, too
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

type Params struct {
	Mode  string `json:"mode"`
	Debug bool   `json:"debug"`
}

type NameValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Response struct {
	Version  Version         `json:"version"`
	Metadata []NameValuePair `json:"metadata,omitempty"`
}

// TODO Perhaps replace with https://github.com/go-playground/validator
func ValidateSource(source Source) error {
	if source.URI == "" {
		return fmt.Errorf("source.uri must not be empty")
	}

	if source.Path == "" {
		return fmt.Errorf("source.path must not be empty")
	}

	return nil
}
