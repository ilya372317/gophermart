package storage

import (
	"context"
	"testing"

	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderFields struct {
	number int `db:"number"`
}

func TestDBStorage_HasOrderByNumber(t *testing.T) {
	tests := []struct {
		name     string
		fields   []orderFields
		argument int
		want     bool
	}{
		{
			name: "success case",
			fields: []orderFields{
				{
					number: 123,
				},
				{
					number: 321,
				},
			},
			argument: 123,
			want:     true,
		},
		{
			name: "not found case",
			fields: []orderFields{
				{
					number: 321,
				},
				{
					number: 567,
				},
			},
			argument: 123,
			want:     false,
		},
		{
			name:     "not found in empty storage case",
			fields:   nil,
			argument: 123,
			want:     false,
		},
	}
	ctx := context.Background()
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOrdersTable(ctx, t)
			fillOrdersTable(ctx, t, tt.fields, userID)
			strg := New(db)
			res, err := strg.HasOrderByNumber(ctx, tt.argument)
			require.NoError(t, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestDBStorage_HasOrderByNumberAndUserID(t *testing.T) {
	tests := []struct {
		name               string
		fields             []orderFields
		argNumber          int
		isArgExistedUserID bool
		want               bool
	}{
		{
			name: "success case",
			fields: []orderFields{
				{
					number: 123,
				},
			},
			argNumber:          123,
			isArgExistedUserID: true,
			want:               true,
		},
		{
			name: "wrong userID case",
			fields: []orderFields{
				{
					number: 123,
				},
			},
			argNumber:          123,
			isArgExistedUserID: false,
			want:               false,
		},
		{
			name:               "wrong userID and number case",
			fields:             nil,
			argNumber:          123,
			isArgExistedUserID: false,
			want:               false,
		},
	}
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		clearOrdersTable(ctx, t)
		fillOrdersTable(ctx, t, tt.fields, userID)
		t.Run(tt.name, func(t *testing.T) {
			repo := New(db)
			argUserID := userID
			if !tt.isArgExistedUserID {
				argUserID = 0
			}
			got, err := repo.HasOrderByNumberAndUserID(ctx, tt.argNumber, argUserID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func fillOrdersTable(ctx context.Context, t *testing.T, fields []orderFields, userID uint) {
	t.Helper()
	orders := make([]entity.Order, 0, len(fields))
	for _, f := range fields {
		orders = append(orders, entity.Order{
			Number: f.number,
			UserID: userID,
		})
	}

	if len(fields) > 0 {
		_, err := db.NamedExecContext(ctx,
			"INSERT INTO orders (user_id, number) VALUES (:user_id, :number)", orders)
		require.NoError(t, err)
	}
}

func clearOrdersTable(ctx context.Context, t *testing.T) {
	t.Helper()
	_, err := db.ExecContext(ctx, "DELETE FROM orders")
	require.NoError(t, err)
}

func createTestUser(ctx context.Context, t *testing.T) uint {
	t.Helper()
	_, err := db.ExecContext(ctx, "INSERT INTO users (login, password) VALUES ('test', 'test')")
	require.NoError(t, err)
	var id uint
	err = db.Get(&id, "SELECT MAX(id) FROM users")
	require.NoError(t, err)

	return id
}

func TestDBStorage_SaveOrder(t *testing.T) {
	tests := []struct {
		name     string
		argument int
		fields   []orderFields
		wantErr  bool
	}{
		{
			name:     "success case with filled storage",
			argument: 123,
			fields: []orderFields{
				{
					number: 321,
				},
				{
					number: 567,
				},
				{
					number: 958,
				},
			},
			wantErr: false,
		},
		{
			name:     "negative already exists case",
			argument: 123,
			fields: []orderFields{
				{
					number: 123,
				},
			},
			wantErr: true,
		},
		{
			name:     "success case with empty storage",
			argument: 123,
			fields:   nil,
			wantErr:  false,
		},
	}
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOrdersTable(ctx, t)
			fillOrdersTable(ctx, t, tt.fields, userID)
			order := entity.Order{
				Number: tt.argument,
				UserID: userID,
			}
			repo := New(db)
			err := repo.SaveOrder(ctx, &order)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			lastSavedOrder := new(entity.Order)
			err = db.Get(lastSavedOrder, "SELECT * FROM orders WHERE id = (SELECT MAX(id) FROM orders)")
			require.NoError(t, err)

			assert.Equal(t, lastSavedOrder.Number, tt.argument)
			assert.Equal(t, lastSavedOrder.UserID, userID)
			assert.Equal(t, lastSavedOrder.Status, "NEW")
			assert.False(t, lastSavedOrder.Accrual.Valid)
		})
	}
}
