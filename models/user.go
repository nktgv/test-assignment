package models

type User struct {
	ID         uint64
	Capacity   int
	RatePerSec int
	Tokens     int
}
