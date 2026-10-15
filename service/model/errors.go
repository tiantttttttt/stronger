package model

import "errors"

var (
	ErrInvalidToken      = errors.New("invalid access token")
	ErrNotFound          = errors.New("workout not found")
	ErrProfileIncomplete = errors.New("profile incomplete")
	ErrTerminalState     = errors.New("workout is already in a terminal state")
)
