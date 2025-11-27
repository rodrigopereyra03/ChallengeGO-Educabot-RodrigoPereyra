package errors

import "errors"

var (
	ErrExternalService = errors.New("external service error")
	ErrTimeout         = errors.New("external service timeout")
	ErrInvalidResponse = errors.New("invalid response from external API")
)
