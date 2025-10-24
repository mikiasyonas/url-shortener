package domain

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrInvalidURL       = errors.New("invalid URL")
	ErrInvalidShortCode = errors.New("invalid short code")
	ErrShortCodeTaken   = errors.New("short code already taken")
)
