package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Add newline for terminal readability
	js = append(js, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {

		// other error - no body
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// syntax error - malformed JSON
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body has malformed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body has malformed JSON")

		// type error - incorrect JSON
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// size error - body is too big
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not exceed %d bytes", maxBytesError.Limit)

		// decode error - n.b. no distinct error type so we have to match to this string
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// operator error - decode received non-nil pointer
		case errors.As(err, *invalidUnmarshalError):
			panic(err)

		// default error - everything else
		default:
			return err
		}
	}

	return nil
}

func (app *application) getFactions() []int {
	return []int{1, 2}
}

func (app *application) getRegions() []string {
	return []string{"us", "eu"}
}

func (app *application) getServers() []string {
	return app.stores.Servers.GetAll()
}

func intToGold(i int) string {
	gold := i / 10_000
	silver := (i % 10_000) / 100
	copper := (i % 10_000) % 100

	switch {
	case gold >= 1:
		return fmt.Sprintf("%vg %vs %vc", gold, silver, copper)
	case silver >= 1:
		return fmt.Sprintf("%vs %vc", silver, copper)
	}

	return fmt.Sprintf("%vc", copper)
}
