package storage

import (
	"context"
	"database/sql"
	"errors"
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
	if err := d.db.QueryRowxContext(ctx,
		"SELECT id, login, password, created_at, updated_at FROM users WHERE id = $1", id).
		StructScan(user); err != nil {
		return nil, fmt.Errorf("failed get user by id: %w", err)
	}
	return user, nil
}

func (d *DBStorage) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	user := &entity.User{}
	if err := d.db.QueryRowxContext(ctx,
		"SELECT id, login, password, created_at, updated_at FROM users WHERE login = $1", login).
		StructScan(user); err != nil {
		return nil, fmt.Errorf("failed find user by id: %w", err)
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

func (d *DBStorage) HasUser(ctx context.Context, login string) (bool, error) {
	user := entity.User{}
	err := d.db.
		QueryRowxContext(ctx, "SELECT id, login, password FROM users WHERE login = $1", login).
		StructScan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed check user with login: %s, %w", login, err)
	}

	return true, nil
}
