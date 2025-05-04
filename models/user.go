package models

type User struct {
	ID         uint64 `db:"id"`
	Capacity   int    `db:"capacity"`
	RatePerSec int    `db:"rate_per_sec"`
	Tokens     int    `db:"tokens"`
}
