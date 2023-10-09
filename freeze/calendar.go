package freeze

import (
	"io"

	"gopkg.in/yaml.v3"
)

type Window struct {
	Name string
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
