package resource

import (
	"fmt"
	"strconv"

	"github.com/orsinium-labs/enum"
)

type Version struct {
	SHA string `json:"sha"`
}

type Source struct {
	URI    string `json:"uri" validate:"required"` // the git resource calls it uri, so we do it, too
	Branch string `json:"branch"`
	Path   string `json:"path" validate:"required,filepath"`
}

type Params struct {
	Mode  Mode `json:"mode"`
	Debug bool `json:"debug"`
}

type Mode enum.Member[string]

var (
	Fuse = Mode{"fuse"}
	Gate = Mode{"gate"}
	Modi = enum.New(Fuse, Gate)
)

func (m *Mode) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))

	if err != nil {
		return err
	}

	parsed := Modi.Parse(unquoted)

	if parsed == nil {
		return fmt.Errorf("%s is not a valid mode, valid ones are %s", string(b), Modi.String())
	}

	*m = *parsed
	return nil
}

type NameValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Response struct {
	Version  Version         `json:"version"`
	Metadata []NameValuePair `json:"metadata,omitempty"`
}
