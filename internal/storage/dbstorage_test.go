package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/ilya372317/gophermart/internal/dbmanager"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/ilya372317/gophermart/internal/logger"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	err := logger.Initialize("")
	if err != nil {
		log.Fatal(err)
		return
	}
	database, pool, resource, err := dbmanager.MakeTestConnection("../../db/migrations")
	if err != nil {
		log.Fatal(err)
		return
	}
	db = database
	m.Run()
	if err = dbmanager.CloseTestConnection(db, pool, resource); err != nil {
		log.Fatal(err)
		return
	}
}

func TestDBStorage_GetUserByID(t *testing.T) {
	type fields struct {
		login    string
		password string
	}
	tests := []struct {
		name     string
		fields   []fields
		argument uint
		want     *entity.User
		wantErr  bool
	}{
		{
			name: "success filled storage case",
			fields: []fields{
				{
					login:    "123",
					password: "test-password",
				},
				{
					login:    "321",
					password: "pass123",
				},
				{
					login:    "543",
					password: "password-543",
				},
			},
			argument: 2,
			want: &entity.User{
				ID:       2,
				Login:    "321",
				Password: "pass123",
			},
			wantErr: false,
		},
		{
			name:     "user not fount case",
			fields:   nil,
			argument: 1,
			want:     nil,
			wantErr:  true,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := make([]entity.User, 0, len(tt.fields))
			for _, f := range tt.fields {
				user := entity.User{
					Login: f.login,
				}
				user.SetPassword(f.password)
				users = append(users, user)
			}

			if len(users) > 0 {
				_, err := db.NamedExecContext(ctx, "INSERT INTO users (login, password) VALUES (:login,:password)", users)
				require.NoError(t, err)
			}

			repo := New(db)
			got, err := repo.GetUserByID(ctx, tt.argument)
			if errors.Is(err, sql.ErrNoRows) {
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			} else {
				require.NoError(t, err)
			}
			if tt.want != nil {
				tt.want.SetPassword(tt.want.Password)
			}
			assert.Equal(t, tt.want, got)
			err = clearUsersTable(ctx)
			require.NoError(t, err)
		})
	}
}

func clearUsersTable(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "DELETE FROM users")
	if err != nil {
		return fmt.Errorf("failed clear users table")
	}

	return nil
}
