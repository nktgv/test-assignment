package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	// sqlx.
	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sqlx.DB
}

func New(username string, password string, host string, port int, dbName string) (*Storage, error) {
	const op = "storage.postgres.New"

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbName)

	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) Close() error {
	return s.DB.Close()
}
