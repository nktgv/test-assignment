package models

import "time"

type Backend struct {
	ID          uint64
	Url         string
	IsAlive     bool
	ActiveConns int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
