package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ilya372317/gophermart/internal/entity"
)

func (d *DBStorage) HasOrderByNumber(ctx context.Context, number int) (bool, error) {
	err := d.db.QueryRowxContext(ctx, "SELECT id FROM orders WHERE number = $1", number).
		StructScan(&entity.Order{})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed check has order by login: %w", err)
	}
	return true, nil
}

func (d *DBStorage) HasOrderByNumberAndUserID(ctx context.Context, number int, userID uint) (bool, error) {
	err := d.db.QueryRowxContext(ctx,
		"SELECT id FROM orders WHERE number = $1 AND user_id = $2", number, userID).
		StructScan(&entity.Order{})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed check has order by number and user id: %w", err)
	}

	return true, nil
}

func (d *DBStorage) SaveOrder(ctx context.Context, order *entity.Order) error {
	_, err := d.db.NamedExecContext(ctx,
		"INSERT INTO orders (user_id, number) VALUES (:user_id, :number)", order)
	if err != nil {
		return fmt.Errorf("failed save order: %w", err)
	}

	return nil
}
