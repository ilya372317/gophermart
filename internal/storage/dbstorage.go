package storage

import (
	"context"
	"fmt"

	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/jmoiron/sqlx"
)

type DBStorage struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *DBStorage {
	return &DBStorage{db: db}
}

func (d *DBStorage) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	user := &entity.User{}
	if err := d.db.QueryRowContext(ctx, "SELECT id, login, password FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Login, &user.Password); err != nil {
		return nil, fmt.Errorf("failed get user by id: %w", err)
	}
	return user, nil
}

func (d *DBStorage) SaveUser(ctx context.Context, user entity.User) error {
	_, err := d.db.ExecContext(
		ctx,
		"INSERT INTO users (login, password) VALUES ($1,$2)",
		user.Login, user.Password,
	)

	if err != nil {
		return fmt.Errorf("failed save user: %w", err)
	}

	return nil
}
