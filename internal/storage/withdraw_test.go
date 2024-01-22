package storage

import (
	"context"
	"testing"

	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func clearWithdrawalTable(ctx context.Context, t *testing.T) {
	t.Helper()
	_, err := db.ExecContext(ctx, "DELETE FROM withdrawals")
	require.NoError(t, err)
}
