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
	var e *Error

	if nil == err {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}

	return EINTERNAL
}

func ErrorMessage(err error) string {
	var e *Error

	if nil == err {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}

	return "internal error"
}

func ErrorDetails(err error) map[string]string {
	var e *Error

	if nil == err {
		return nil
	} else if errors.As(err, &e) {
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
