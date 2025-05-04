package models

import "time"

type Backend struct {
	ID          uint64    `db:"id"`
	Url         string    `db:"url"`
	IsAlive     bool      `db:"is_alive"`
	ActiveConns int       `db:"active_conns"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
