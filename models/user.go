package models

type User struct {
	ID          uint64 `db:"id"           json:"client_id"`
	Capacity    int    `db:"capacity"     json:"capacity"`
	RatePerSec  int    `db:"rate_per_sec" json:"rate_per_sec"`
	Tokens      int    `db:"tokens"       json:"tokens"`
	LastUpdated string `db:"last_updated"`
}
