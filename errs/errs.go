package errs

import (
	"errors"
	"fmt"
)

const (
	ECONFLICT             = "conflict"
	EFORBIDDEN            = "forbidden"
	EINTERNAL             = "internal"
	EINVALID              = "invalid"
	ENOTACCEPTABLE        = "no_acceptable"
	ENOTFOUND             = "not_found"
	ENOTIMPLEMENTED       = "not_implemented"
	ETOOMANYREQUESTS      = "too_many_requests"
	EUNPROCESSABLECONTENT = "unprocessable_content"
	EUNAUTHORIZED         = "unauthorized"
)

type Error struct {
	Code    string
	Message string
	Details map[string]string
}

func (e *Error) Error() string {
	return fmt.Sprintf("webservice error: code=%s message=%s", e.Code, e.Message)
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Code
	}

	return EINTERNAL
}

func Is(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Message
	}

	return "internal error"
}

func ErrorDetails(err error) map[string]string {
	if err == nil {
		return nil
	}
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Details
	}

	return nil
}

func Errorf(code, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
