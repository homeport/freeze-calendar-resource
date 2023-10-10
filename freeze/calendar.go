package freeze

import (
	"io"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Window struct {
	Name  string    `yaml:"name" validate:"required"`
	Start time.Time `yaml:"starts_at" validate:"required"`
	End   time.Time `yaml:"ends_at" validate:"required"`
	Scope []string  `yaml:"scope"`
}

type Calendar struct {
	Windows []Window `yaml:"freeze_calendar" validate:"omitempty,dive"`
}

func LoadCalendar(reader io.Reader) (*Calendar, error) {
	var calendar Calendar
	err := yaml.NewDecoder(reader).Decode(&calendar)

	if err != nil {
		return nil, err
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(calendar)

	if err != nil {
		return nil, err
	}

	return &calendar, nil
}
