package freeze

import (
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

type Window struct {
	Name  string    `yaml:"name"`
	Start time.Time `yaml:"starts_at"`
	End   time.Time `yaml:"ends_at"`
	Scope []string  `yaml:"scope"`
}

type Calendar struct {
	Windows []Window `yaml:"freeze_calendar"`
}

func LoadCalendar(reader io.Reader) (*Calendar, error) {
	var calendar Calendar
	err := yaml.NewDecoder(reader).Decode(&calendar)

	if err != nil {
		return nil, err
	}

	return &calendar, nil
}
