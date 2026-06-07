package apperror

import "errors"

const CodeInternal int32 = 50000

type Error struct {
	Code    int32
	Message string
	Err     error
}

func New(code int32, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return e.Message
	}
	return e.Message + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func From(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func OrInternal(err error) *Error {
	if appErr, ok := From(err); ok {
		return appErr
	}
	return New(CodeInternal, "internal server error", err)
}
