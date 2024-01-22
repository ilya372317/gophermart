package storage

import (
	"context"
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
