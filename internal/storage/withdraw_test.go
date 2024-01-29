package storage

import (
	"context"
	"testing"

	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type withdrawalFields struct {
	order int
	sum   int
}

func TestDBStorage_SaveWithdrawal(t *testing.T) {
	tests := []struct {
		name        string
		argument    entity.Withdrawal
		userIDExist bool
		wantErr     bool
	}{
		{
			name: "success save case",
			argument: entity.Withdrawal{
				OrderNumber: 123,
				Sum:         10,
			},
			userIDExist: true,
			wantErr:     false,
		},
		{
			name: "invalid user id case",
			argument: entity.Withdrawal{
				OrderNumber: 10,
				Sum:         20,
			},
			userIDExist: false,
			wantErr:     true,
		},
	}
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearWithdrawalTable(ctx, t)
			repo := New(db)
			if tt.userIDExist {
				tt.argument.UserID = userID
			}
			err := repo.SaveWithdrawal(ctx, tt.argument)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			savedWithdrawal := entity.Withdrawal{}
			err = db.Get(&savedWithdrawal,
				"SELECT * FROM withdrawals WHERE id = (SELECT MAX(id) FROM withdrawals)")
			require.NoError(t, err)
			assert.Equal(t, tt.argument.Sum, savedWithdrawal.Sum)
			assert.Equal(t, userID, savedWithdrawal.UserID)
			assert.Equal(t, tt.argument.OrderNumber, savedWithdrawal.OrderNumber)
		})
	}
}

func TestDBStorage_GetWithdrawalListByUserID(t *testing.T) {
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	tests := []struct {
		name   string
		fields []withdrawalFields
		want   []entity.Withdrawal
	}{
		{
			name:   "empty case",
			fields: nil,
			want:   nil,
		},
		{
			name: "filled case",
			fields: []withdrawalFields{
				{
					order: 123,
					sum:   10,
				},
				{
					order: 321,
					sum:   20,
				},
				{
					order: 456,
					sum:   30,
				},
			},
			want: []entity.Withdrawal{
				{
					OrderNumber: 123,
					UserID:      userID,
					Sum:         10,
				},
				{
					OrderNumber: 321,
					UserID:      userID,
					Sum:         20,
				},
				{
					OrderNumber: 456,
					UserID:      userID,
					Sum:         30,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearOrdersTable(ctx, t)
			fillWithdrawalTable(ctx, t, tt.fields, userID)
			repo := New(db)
			gotSlice, err := repo.GetWithdrawalListByUserID(ctx, userID)
			require.NoError(t, err)

			for _, w := range tt.want {
				var got entity.Withdrawal
				if len(gotSlice) == 0 {
					t.Errorf("inserted withdrawals count not equal to got")
					return
				}
				got, gotSlice = gotSlice[0], gotSlice[1:]
				assert.Equal(t, w.Sum, got.Sum)
				assert.Equal(t, w.OrderNumber, got.OrderNumber)
				assert.Equal(t, w.UserID, got.UserID)
			}
		})
	}
}

func fillWithdrawalTable(ctx context.Context, t *testing.T, fields []withdrawalFields, userID uint) {
	t.Helper()
	if len(fields) == 0 {
		return
	}
	withdrawals := make([]entity.Withdrawal, 0, len(fields))
	for _, f := range fields {
		withdrawal := entity.Withdrawal{
			OrderNumber: f.order,
			UserID:      userID,
			Sum:         f.sum,
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	_, err := db.NamedExecContext(ctx,
		"INSERT  INTO withdrawals (order_number, sum, user_id) VALUES (:order_number, :sum, :user_id)",
		withdrawals)
	require.NoError(t, err)
}

func clearWithdrawalTable(ctx context.Context, t *testing.T) {
	t.Helper()
	_, err := db.ExecContext(ctx, "DELETE FROM withdrawals")
	require.NoError(t, err)
}

func TestDBStorage_GetWithdrawalSumByUserID(t *testing.T) {
	tests := []struct {
		name   string
		fields []withdrawalFields
		want   int
	}{
		{
			name: "success case with multiply items",
			fields: []withdrawalFields{
				{
					order: 123,
					sum:   10,
				},
				{
					order: 321,
					sum:   20,
				},
				{
					order: 456,
					sum:   20,
				},
			},
			want: 50,
		},
		{
			name: "success case with one item",
			fields: []withdrawalFields{
				{
					order: 123,
					sum:   10,
				},
			},
			want: 10,
		},
		{
			name:   "empty case",
			fields: nil,
			want:   0,
		},
	}
	ctx := context.Background()
	clearUsersTable(ctx, t)
	userID := createTestUser(ctx, t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearWithdrawalTable(ctx, t)
			fillWithdrawalTable(ctx, t, tt.fields, userID)
			repo := New(db)
			got, err := repo.GetWithdrawalSumByUserID(ctx, userID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
