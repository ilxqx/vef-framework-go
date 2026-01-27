package auth

import "errors"

var (
	ErrMissingToken = errors.New("missing token")
	ErrInvalidToken = errors.New("invalid token")
)
