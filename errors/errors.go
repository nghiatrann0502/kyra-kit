package errors

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
)

type Error struct {
	code    ErrorCode
	message string
	cause   error
	stack   []Frame
	fields  map[string]any
}

func New(code ErrorCode, message string) *Error {
	return &Error{
		code:    code,
		message: message,
		stack:   captureStack(3),
		fields:  make(map[string]any),
	}
}

func Newf(code ErrorCode, format string, args ...any) *Error {
	return New(code, fmt.Sprintf(format, args...))
}

func WrapInternalErr(err error) *Error {
	if err == nil {
		return nil
	}

	return Wrap(err, GetCode(err), http.StatusText(http.StatusInternalServerError))
}

func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return nil
	}
	var stack []Frame

	var leaf *Error
	if errors.As(err, &leaf) {
		stack = leaf.stack
	} else {
		stack = captureStack(3)
	}

	e := &Error{
		code:    code,
		message: message,
		cause:   err,
		stack:   stack, // Skip Wrap and caller.
		fields:  make(map[string]any),
	}

	// If we're wrapping an Error, copy its fields
	var domainErr *Error
	if errors.As(err, &domainErr) {
		maps.Copy(e.fields, domainErr.fields)
	}

	return e
}

func Wrapf(err error, code ErrorCode, format string, args ...any) *Error {
	if err == nil {
		return nil
	}
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func (err *Error) Error() string {
	if err.cause != nil {
		return fmt.Sprintf("%d: %s: %v", err.code, err.message, err.cause)
	}

	return fmt.Sprintf("%d: %s", err.code, err.message)
}

func (err *Error) Message() string {
	return err.message
}

func (err *Error) Cause() error {
	return err.cause
}

func GetCode(err error) ErrorCode {
	var appErr *Error
	if As(err, &appErr) {
		return appErr.code
	}
	return ErrCodeInternal
}

func (err *Error) StackTrace() string {
	var sb strings.Builder
	for i, frame := range err.stack {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
	}
	return sb.String()
}

func (err *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return err.code == t.code
}

func (err *Error) Unwrap() error {
	return err.cause
}

func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

func (err *Error) WithField(field string, data any) *Error {
	if err.fields == nil {
		err.fields = make(map[string]any)
	}

	err.fields[field] = data
	return err
}
