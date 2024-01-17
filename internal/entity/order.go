package entity

import (
	"database/sql"
	"time"
)

type Order struct {
	CreatedAT time.Time     `db:"created_at"`
	UpdatedAT time.Time     `db:"updated_at"`
	Status    string        `db:"status"`
	Number    string        `db:"number"`
	Accrual   sql.NullInt64 `db:"accrual"`
	ID        uint          `db:"id"`
	UserID    uint          `db:"user_id"`
}
