package entity

import "time"

type Withdrawal struct {
	CreatedAT   time.Time `db:"created_at"`
	UpdatedAT   time.Time `db:"updated_at"`
	OrderNumber string    `db:"order_number"`
	ID          uint      `db:"id"`
	UserID      uint      `db:"user_id"`
	Sum         int       `db:"sum"`
}
