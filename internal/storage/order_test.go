package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderFields struct {
	number  int `db:"number"`
	accrual sql.NullInt64
	status  string
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
		if len(f.status) == 0 {
			f.status = "NEW"
		}
		orders = append(orders, entity.Order{
			Number:  f.number,
			UserID:  userID,
			Accrual: f.accrual,
			Status:  f.status,
		})
	}

	if len(fields) > 0 {
		_, err := db.NamedExecContext(ctx,
			"INSERT INTO orders (user_id, number, accrual, status) VALUES (:user_id, :number, :accrual, :status)",
			orders)
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

func TestDBStorage_GetOrderByNumber(t *testing.T) {
	type want struct {
		number  int
		status  string
		accrual sql.NullInt64
	}
	tests := []struct {
		name     string
		argument int
		fields   []orderFields
		want     want
		wantErr  bool
	}{
		{
			name:     "success case",
			argument: 123,
			fields: []orderFields{
				{
					number: 123,
					accrual: sql.NullInt64{
						Int64: 123,
						Valid: true,
					},
					status: "INVALID",
				},
			},
			want: want{
				number: 123,
				status: "INVALID",
				accrual: sql.NullInt64{
					Int64: 123,
					Valid: true,
				},
			},
			wantErr: false,
		},
	}
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOrdersTable(ctx, t)
			fillOrdersTable(ctx, t, tt.fields, userID)
			repo := New(db)
			got, err := repo.GetOrderByNumber(ctx, tt.argument)
			if tt.wantErr {
				require.ErrorIs(t, err, sql.ErrNoRows)
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.number, got.Number)
			assert.Equal(t, tt.want.accrual, got.Accrual)
			assert.Equal(t, tt.want.status, got.Status)
		})
	}
}

func TestDBStorage_UpdateOrderStatusByNumber(t *testing.T) {
	type argument struct {
		number int
		status string
	}

	tests := []struct {
		name     string
		field    orderFields
		argument argument
		wantErr  bool
	}{
		{
			name: "success case",
			field: orderFields{
				number: 123,
			},
			argument: argument{
				number: 123,
				status: "INVALID",
			},
			wantErr: false,
		},
		{
			name: "invalid status case",
			field: orderFields{
				number: 123,
			},
			argument: argument{
				number: 123,
				status: "INVALID_NAME",
			},
			wantErr: true,
		},
		{
			name: "no rows affected case",
			field: orderFields{
				number: 123,
			},
			argument: argument{
				number: 321,
				status: "PROCESSED",
			},
			wantErr: true,
		},
		{
			name: "update status to PROCESSING",
			field: orderFields{
				number: 123,
			},
			argument: argument{
				number: 123,
				status: "PROCESSING",
			},
			wantErr: false,
		},
		{
			name: "update status to INVALID",
			field: orderFields{
				number: 123,
			},
			argument: argument{
				number: 123,
				status: "INVALID",
			},
			wantErr: false,
		},
		{
			name: "update status to PROCESSED",
			field: orderFields{
				number: 123,
			},
			argument: argument{
				number: 123,
				status: "PROCESSED",
			},
			wantErr: false,
		},
	}
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOrdersTable(ctx, t)
			fields := make([]orderFields, 0)
			fields = append(fields, tt.field)
			fillOrdersTable(ctx, t, fields, userID)
			repo := New(db)
			err := repo.UpdateOrderStatusByNumber(ctx, tt.argument.number, tt.argument.status)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			orderAfterUpdate, err := repo.GetOrderByNumber(ctx, tt.field.number)
			require.NoError(t, err)
			assert.Equal(t, tt.argument.number, orderAfterUpdate.Number)
			assert.Equal(t, tt.argument.status, orderAfterUpdate.Status)
		})
	}
}

func TestDBStorage_UpdateOrderAccrualByNumber(t *testing.T) {
	type argument struct {
		number  int
		accrual sql.NullInt64
	}
	tests := []struct {
		name     string
		field    orderFields
		wantErr  bool
		argument argument
	}{
		{
			name: "success case",
			field: orderFields{
				number: 123,
				accrual: sql.NullInt64{
					Int64: 10,
					Valid: true,
				},
			},
			wantErr: false,
			argument: argument{
				number: 123,
				accrual: sql.NullInt64{
					Int64: 20,
					Valid: true,
				},
			},
		},
		{
			name: "no rows was updated case",
			field: orderFields{
				number: 321,
				accrual: sql.NullInt64{
					Int64: 10,
					Valid: true,
				},
			},
			wantErr: true,
			argument: argument{
				number: 123,
				accrual: sql.NullInt64{
					Int64: 123,
					Valid: true,
				},
			},
		},
		{
			name: "update order accrual from null to value",
			field: orderFields{
				number: 123,
			},
			wantErr: false,
			argument: argument{
				number: 123,
				accrual: sql.NullInt64{
					Int64: 10,
					Valid: true,
				},
			},
		},
		{
			name: "update order accrual from value to null",
			field: orderFields{
				number: 123,
				accrual: sql.NullInt64{
					Int64: 123,
					Valid: true,
				},
			},
			wantErr: false,
			argument: argument{
				number: 123,
			},
		},
	}
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOrdersTable(ctx, t)
			fields := make([]orderFields, 0)
			fields = append(fields, tt.field)
			fillOrdersTable(ctx, t, fields, userID)
			repo := New(db)
			err := repo.UpdateOrderAccrualByNumber(ctx, tt.argument.number, tt.argument.accrual)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			got := getLastOrder(ctx, t)
			assert.Equal(t, tt.argument.number, got.Number)
			assert.Equal(t, tt.argument.accrual, got.Accrual)
		})
	}
}

func getLastOrder(ctx context.Context, t *testing.T) *entity.Order {
	t.Helper()
	order := &entity.Order{}
	err := db.GetContext(ctx, order, "SELECT * FROM orders WHERE id = (SELECT MAX(id) FROM orders)")
	require.NoError(t, err)
	return order
}
