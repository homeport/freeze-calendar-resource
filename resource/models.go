package resource

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/orsinium-labs/enum"
)

type Request struct {
	Source Source `json:"source" validate:"required"`
}

type Params struct {
	Mode   Mode     `json:"mode" validate:"required"`
	Scope  []string `json:"scope"`
	Runway Duration `json:"runway"`
}

type Mode enum.Member[string]

var (
	Fuse = Mode{"fuse"}
	Gate = Mode{"gate"}
	Modi = enum.New(Fuse, Gate)
)

func (m *Mode) UnmarshalJSON(b []byte) error {
	rawMode := string(b)
	unquoted, err := strconv.Unquote(rawMode)

	if err != nil {
		return fmt.Errorf("unable to unquote raw mode '%s': %w", rawMode, err)
	}

	parsed := Modi.Parse(unquoted)

	if parsed == nil {
		return fmt.Errorf("%s is not a valid mode, valid ones are %s", string(b), Modi.String())
	}

	*m = *parsed
	return nil
}

type Version struct {
	SHA string `json:"sha"`
}

type NameValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Source struct {
	URI        string `json:"uri" validate:"required"` // the git resource calls it uri, so we do it, too
	PrivateKey string `json:"private_key"`
	KnownHosts string `json:"known_hosts"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Branch     string `json:"branch"`
	Path       string `json:"path" validate:"required,filepath"`
	Debug      bool   `json:"debug"`
}

func (source Source) Auth() (auth transport.AuthMethod, err error) {
	if source.Username != "" && source.Password != "" {
		auth = &http.BasicAuth{
			Username: source.Username,
			Password: source.Password,
		}
	}

	if len(source.PrivateKey) != 0 {
		if auth != nil {
			return nil, errors.New("both private_key and {username, password} are set, but only one of these is allowed")
		}

		auth, err = ssh.NewPublicKeys(
			// there seems to be no good library for parsing git URLs; this is the poor man's approach.
			strings.SplitN(source.URI, "@", 2)[0],
			[]byte(source.PrivateKey),
			"",
		)

		if err != nil {
			return nil, fmt.Errorf("could not create public key pair: %w", err)
		}

		if len(source.KnownHosts) != 0 {
			f, err := os.CreateTemp("", "known-hosts-")

			if err != nil {
				return nil, fmt.Errorf("could not create temporary file for known_hosts: %w", err)
			}

			_, err = f.Write([]byte(source.KnownHosts))

			if err != nil {
				return nil, fmt.Errorf("could not write contents of SSH_KNOWN_HOSTS to %s: %w", f.Name(), err)
			}

			err = os.Setenv("SSH_KNOWN_HOSTS", f.Name()) // simpler than dealing with ssh.HostKeyCallback

			if err != nil {
				return nil, fmt.Errorf("could not change env var SSH_KNOWN_HOSTS to point to %s: %w", f.Name(), err)
			}
		}
	}

	return auth, nil
}
