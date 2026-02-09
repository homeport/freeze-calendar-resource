package freeze

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"go.yaml.in/yaml/v3"
)

type Window struct {
	Name  string    `yaml:"name" validate:"required"`
	Start time.Time `yaml:"starts_at" validate:"required"`
	End   time.Time `yaml:"ends_at" validate:"required,gtcsfield=Start"`
	Scope []string  `yaml:"scope,omitempty"`
}

func (w Window) String() (result string) {
	result = fmt.Sprintf("%s from %s to %s", w.Name, w.Start, w.End)

	if len(w.Scope) > 0 {
		result += fmt.Sprintf("; scope: %s", strings.Join(w.Scope, ", "))
	}

	return
}

type Calendar struct {
	Windows []Window `yaml:"freeze_calendar" validate:"omitempty,dive"`
}

func LoadCalendar(reader io.Reader) (*Calendar, error) {
	var calendar Calendar
	err := yaml.NewDecoder(reader).Decode(&calendar)

	if err != nil {
		return nil, fmt.Errorf("unable to build decoder: %w", err)
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(calendar)

	if err != nil {
		return nil, fmt.Errorf("unable to build validator: %w", err)
	}

	return &calendar, nil
}
