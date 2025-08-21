package errors

import "errors"

type ErrorCode int

const (
	ErrCodeForbidden ErrorCode = 1000

	ErrCodeUnauthorized ErrorCode = 2000

	ErrCodeInvalidInput  ErrorCode = 4000
	ErrCodeNotFound      ErrorCode = 4040
	ErrCodeAlreadyExists ErrorCode = 4090

	ErrCodeUnknown  ErrorCode = 5000
	ErrCodeInternal ErrorCode = 5001
	ErrCodeDatabase ErrorCode = 5002
	ErrCodeExternal ErrorCode = 5003
)

var ErrNotFound = errors.New("not found")
