package resource

type Request struct {
	Version Version `json:"version,omitempty"`
	Source  Source  `json:"source" validate:"required"`
}

type Version struct {
	SHA string `json:"sha"`
}

type Source struct {
	URI    string `json:"uri" validate:"required"` // the git resource calls it uri, so we do it, too
	Branch string `json:"branch"`
	Path   string `json:"path" validate:"required,filepath"`
}

type Response struct {
	Version  Version         `json:"version"`
	Metadata []NameValuePair `json:"metadata,omitempty"`
}

type NameValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
