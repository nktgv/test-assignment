package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"http-load-balancer/models"
)

type UserRepository interface {
	GetAll() ([]models.User, error)
	GetByID(id uint64) (models.User, error)
	Create(user *models.User) (*models.User, error)
	UpdateTokens(id uint64, tokens int) (bool, error)
	UpdateCapacity(id uint64, capacity int) (bool, error)
	UpdateRatePerSec(id uint64, ratePerSecond int) (bool, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]models.User, error) {
	const op = "userRepository.GetAll"

	users := make([]models.User, 0)
	err := r.db.Select(&users, `SELECT * FROM user`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return users, nil
}

func (r *userRepository) GetByID(id uint64) (models.User, error) {
	const op = "userRepository.GetByID"

	user := models.User{}
	err := r.db.Get(&user, `SELECT * FROM user WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, err)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

func (r *userRepository) Create(user *models.User) (*models.User, error) {
	const op = "userRepository.Create"

	var userID uint64
	err := r.db.Get(&userID,
		`
			INSERT INTO user (capacity, rate_per_sec, tokens)
			VALUES ($1, $2, $3)
			RETURNING id
	`,
		user.Capacity, user.RatePerSec, user.Tokens,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	user.ID = userID
	return user, nil
}

func (r *userRepository) UpdateTokens(id uint64, tokens int) (bool, error) {
	const op = "userRepository.UpdateTokens"

	res, err := r.db.Exec(`
		UPDATE users SET tokens=$1 WHERE id=$2
	`, tokens, id)
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

func (r *userRepository) UpdateCapacity(id uint64, capacity int) (bool, error) {
	const op = "userRepository.UpdateCapacity"

	res, err := r.db.Exec(`
		UPDATE users SET capacity=$1 WHERE id=$2
	`, capacity, id)
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

func (r *userRepository) UpdateRatePerSec(id uint64, ratePerSecond int) (bool, error) {
	const op = "userRepository.UpdateRatePerSec"

	res, err := r.db.Exec(`
		UPDATE users SET rate_per_sec=$1 WHERE id=$2
	`, ratePerSecond, id)
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
