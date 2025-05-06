package repository

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrBackendNotFound  = errors.New("backend not found")
	ErrNoActiveBackends = errors.New("no active backends")
)
