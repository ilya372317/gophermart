package entity

import (
	"database/sql"
	"time"
)

const (
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
	StatusNew        = "NEW"
)

type Order struct {
	CreatedAT time.Time     `db:"created_at"`
	UpdatedAT time.Time     `db:"updated_at"`
	Status    string        `db:"status"`
	Number    int           `db:"number"`
	Accrual   sql.NullInt64 `db:"accrual"`
	ID        uint          `db:"id"`
	UserID    uint          `db:"user_id"`
}
