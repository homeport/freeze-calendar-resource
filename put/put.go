package put

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/resource"
)

type Request struct {
	resource.Source `json:"source"`
	Params          resource.Params `json:"params" validate:"required"`
}

type Response struct {
	Version  resource.Version         `json:"version"`
	Metadata []resource.NameValuePair `json:"metadata,omitempty"`
}

func Put(ctx context.Context, req io.Reader, resp, log io.Writer, source string) error {
	var request Request
	err := json.NewDecoder(req).Decode(&request)

	if err != nil {
		return fmt.Errorf("unable to build decoder: %w", err)
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

	if err != nil {
		return fmt.Errorf("unable to build validator: %w", err)
	}

	fmt.Fprintln(log, "no-op")

	response := Response{} // no version as we don't put anything

	err = json.NewEncoder(resp).Encode(response)

	if err != nil {
		return fmt.Errorf("unable to encode response as JSON: %w", err)
	}

	return nil
}
