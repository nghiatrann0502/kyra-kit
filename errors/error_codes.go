package errors

import "errors"

type ErrorCode int

const (
	ErrCodeForbidden ErrorCode = 1000

	ErrCodeUnauthorized ErrorCode = 2000

	ErrCodeInvalidInput  ErrorCode = 4000
	ErrCodeNotFound      ErrorCode = 4040
	ErrCodeAlreadyExists ErrorCode = 4090

	ErrCodeInvalidRequest          ErrorCode = 4100
	ErrCodeUnauthorizedClient      ErrorCode = 4101
	ErrCodeAccessDenined           ErrorCode = 4103
	ErrCodeUnsupportedResponseType ErrorCode = 4104
	ErrCodeInvalidScope            ErrorCode = 4105
	ErrCodeTemporarilyUnavailable  ErrorCode = 4106
	ErrCodeLoginRequired           ErrorCode = 4107
	ErrCodeInteractionRequired     ErrorCode = 4108
	ErrCodeInvalidGrant            ErrorCode = 4109

	ErrCodeUnknown     ErrorCode = 5000
	ErrCodeInternal    ErrorCode = 5001
	ErrCodeDatabase    ErrorCode = 5002
	ErrCodeExternal    ErrorCode = 5003
	ErrCOdeServerError ErrorCode = 5004
)

var ErrNotFound = errors.New("not found")
