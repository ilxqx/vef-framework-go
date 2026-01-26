package auth

import "errors"

var (
	ErrMissingToken       = errors.New("missing token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrMissingAuthHeaders = errors.New("missing authentication headers")
	ErrInvalidTimestamp   = errors.New("invalid timestamp")
	ErrRequestExpired     = errors.New("request expired")
	ErrInvalidSignature   = errors.New("invalid signature")
)
