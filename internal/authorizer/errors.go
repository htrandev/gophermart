package authorizer

import "errors"

var (
	ErrUnexpectedMethod = errors.New("unexpected signing method")
	ErrInvalidClaims    = errors.New("invalid token claims")
)
