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

func (d *DBStorage) GetOrderByNumber(ctx context.Context, number int) (*entity.Order, error) {
	order := &entity.Order{}
	err := d.db.GetContext(ctx, order, "SELECT * FROM orders WHERE number = $1", number)
	if err != nil {
		return nil, fmt.Errorf("failed get order by number: %w", err)
	}

	return order, nil
}

func (d *DBStorage) UpdateOrderStatusByNumber(ctx context.Context, number int, status string) error {
	res, err := d.db.ExecContext(ctx, "UPDATE orders SET status = $1 WHERE number = $2", status, number)
	if err != nil {
		return fmt.Errorf("failed update order status by number: %w", err)
	}

	updatedRows, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed get affected rows count: %w", err)
	}

	if updatedRows == 0 {
		return fmt.Errorf("no rows was updated")
	}

	return nil
}

func (d *DBStorage) UpdateOrderAccrualByNumber(ctx context.Context, number int, accrual sql.NullFloat64) error {
	res, err := d.db.ExecContext(ctx, "UPDATE orders SET accrual = $1 WHERE number = $2", accrual, number)
	if err != nil {
		return fmt.Errorf("failed update oreder accrual: %w", err)
	}

	updatedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed get affected rows on order accrual update: %w", err)
	}

	if updatedRows == 0 {
		return fmt.Errorf("no accrual fields of orders was updated")
	}

	return nil
}

func (d *DBStorage) GetOrderListByUserID(ctx context.Context, userID uint) ([]entity.Order, error) {
	orderList := make([]entity.Order, 0)
	err := d.db.SelectContext(ctx, &orderList, "SELECT * FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed get order list by userID: %w", err)
	}

	return orderList, nil
}

func (d *DBStorage) GetOrderListByStatus(ctx context.Context, status string) ([]entity.Order, error) {
	orderList := make([]entity.Order, 0)
	if err := d.db.SelectContext(ctx,
		&orderList, "SELECT * FROM orders WHERE status = $1", status); err != nil {
		return nil, fmt.Errorf("failed get order list by status: %w", err)
	}
	return orderList, nil
}
