package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"http-load-balancer/models"
)

type BackendRepository interface {
	GetAll() ([]models.Backend, error)
	GetActive() ([]models.Backend, error)
	Add(b *models.Backend) (*models.Backend, error)
	SetIsAlive(id uint64, isAlive bool) (bool, error)
}

type backendRepository struct {
	db *sqlx.DB
}

func NewBackendRepository(db *sqlx.DB) BackendRepository {
	return &backendRepository{db: db}
}

func (r *backendRepository) GetAll() ([]models.Backend, error) {
	const op = "BackendRepository.GetAll"

	backends := make([]models.Backend, 0)
	err := r.db.Select(&backends, `SELECT * FROM backend`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return backends, nil
}

func (r *backendRepository) GetActive() ([]models.Backend, error) {
	const op = "BackendRepository.GetAll"

	backends := make([]models.Backend, 0)
	query := `SELECT * FROM backend WHERE is_alive=$1`
	err := r.db.Get(&backends, query, "true")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return backends, nil
}

func (r *backendRepository) Add(b *models.Backend) (*models.Backend, error) {
	const op = "BackendRepository.Add"

	var backendID uint64
	err := r.db.Get(
		&backendID,
		`
			INSERT INTO backend (url, is_alive, created_at, updated_at) 
			VALUES($1, $2, $3, $4) 
			RETURNING id
		`,
		b.Url, b.IsAlive, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	b.ID = backendID
	return b, nil
}

func (r *backendRepository) SetIsAlive(id uint64, isAlive bool) (bool, error) {
	const op = "BackendRepository.SetActive"

	res, err := r.db.Exec(`
		UPDATE backend SET is_alive=$1 WHERE id=$2
	`, isAlive, id)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return false, nil
	}
	return true, nil
}
