package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ilya372317/gophermart/internal/entity"
)

func (d *DBStorage) SaveWithdrawal(ctx context.Context, withdrawal entity.Withdrawal) error {
	_, err := d.db.NamedExecContext(ctx,
		"INSERT INTO withdrawals (order_number, sum, user_id) VALUES (:order_number, :sum, :user_id)",
		withdrawal)
	if err != nil {
		return fmt.Errorf("failed save withdrawal: %w", err)
	}

	return nil
}

func (d *DBStorage) GetWithdrawalListByUserID(ctx context.Context, userID uint) ([]entity.Withdrawal, error) {
	withdrawals := make([]entity.Withdrawal, 0)
	err := d.db.SelectContext(ctx,
		&withdrawals, "SELECT * FROM withdrawals WHERE user_id = $1 ORDER BY created_at", userID)
	if err != nil {
		return nil, fmt.Errorf("failed get withdrawal list from db: %w", err)
	}
	return withdrawals, nil
}

func (d *DBStorage) GetWithdrawalSumByUserID(ctx context.Context, userID uint) (float64, error) {
	var result sql.NullFloat64
	err := d.db.GetContext(ctx, &result, "SELECT SUM(sum) FROM withdrawals WHERE user_id = $1", userID)
	if err != nil {
		return 0, fmt.Errorf("failed get sum of withdrawals by user id: %w", err)
	}

	if result.Valid {
		return result.Float64, nil
	} else {
		return 0, nil
	}
}
