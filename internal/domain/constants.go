package domain

import "errors"

var (
	ErrInvalidSubRepo = errors.New("subscription repository not defined")
	ErrInvalidLogger  = errors.New("logger is not defined")
	ErrInvalidInput   = errors.New("invalid input")
)
