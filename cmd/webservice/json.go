package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/gmr458/receipt-processor-go/errs"
	"github.com/gmr458/receipt-processor-go/money"
)

type envelope map[string]any

func (app *app) sendJSON(w http.ResponseWriter, status int, data any, headers http.Header) {
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(js); err != nil {
		app.logger.Error("failed to write response body: " + err.Error())
	}
}

func (app *app) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	const maxBytes = 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)
	if err != nil {
		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
			maxBytesError         *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxError):
			return errs.Errorf(
				errs.EINVALID,
				"body contains badly-formed JSON (at character %d)",
				syntaxError.Offset,
			)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errs.Errorf(errs.EINVALID, "body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return errs.Errorf(
					errs.EINVALID,
					"body contains incorrect JSON type for field %q",
					unmarshalTypeError.Field,
				)
			}
			return errs.Errorf(
				errs.EINVALID,
				"body contains incorrect JSON type (at character %d)",
				unmarshalTypeError.Offset,
			)

		case errors.Is(err, io.EOF):
			return errs.Errorf(errs.EINVALID, "body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return errs.Errorf(
				errs.EINVALID,
				"body contains unknown key %s",
				fieldName,
			)

		case errors.As(err, &maxBytesError):
			return errs.Errorf(errs.EINVALID, "body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.Is(err, money.ErrInvalidFormat), errors.Is(err, money.ErrOverflow):
			return errs.Errorf(errs.EINVALID, "%s", err.Error())

		case errs.Is(err):
			return err

		case errors.As(err, &invalidUnmarshalError):
			panic(fmt.Sprintf("readJSON: invalid unmarshal target: %v", err))

		default:
			return errs.Errorf(errs.EINTERNAL, "%s", err.Error())
		}
	}

	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errs.Errorf(errs.EINVALID, "body must only contain a single JSON value")
	}

	return nil
}
