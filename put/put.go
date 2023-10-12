package put

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/homeport/freeze-calendar-resource/resource"
)

func Put(ctx context.Context, req io.Reader, resp, log io.Writer, source string) error {
	var request resource.Request
	err := json.NewDecoder(req).Decode(&request)

	if err != nil {
		return err
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(request)

	if err != nil {
		return err
	}

	fmt.Fprintln(log, "no-op")

	response := resource.Response{
		Version: request.Version,
	}

	return json.NewEncoder(resp).Encode(response)
}
