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
		})
		err := clearUsersTable(ctx)
		require.NoError(t, err)
	}
}

func TestDBStorage_SaveUser(t *testing.T) {
	type fields struct {
		login    string
		password string
	}
	type arg struct {
		login    string
		password string
	}

	type want struct {
		login    string
		password string
	}

	tests := []struct {
		name     string
		argument arg
		fields   []fields
		wantErr  bool
		want     want
	}{
		{
			name: "success simple case",
			argument: arg{
				login:    "test",
				password: "test123",
			},
			wantErr: false,
			want: want{
				login:    "test",
				password: "test123",
			},
			fields: []fields{
				{
					login:    "123456",
					password: "sdfsdfsrwer",
				},
			},
		},
		{
			name: "error on duplicate values",
			argument: arg{
				login:    "test",
				password: "123",
			},
			fields: []fields{
				{
					login:    "test",
					password: "123",
				},
			},
			wantErr: true,
			want:    want{},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := db.Beginx()
			require.NoError(t, err)
			existedUsers := make([]entity.User, 0, len(tt.fields))
			for _, field := range tt.fields {
				existedUser := entity.User{
					Login: field.login,
				}
				existedUser.SetPassword(field.password)
				existedUsers = append(existedUsers, existedUser)
			}
			_, err = tx.NamedExecContext(ctx, "INSERT INTO users (login, password) VALUES (:login,:password)", existedUsers)
			require.NoError(t, err)
			err = tx.Commit()
			require.NoError(t, err)
			repo := New(db)
			user := entity.User{
				Login: tt.argument.login,
			}
			user.SetPassword(tt.argument.password)
			err = repo.SaveUser(ctx, user)

			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			expectedUser := &entity.User{
				Login: tt.want.login,
			}
			expectedUser.SetPassword(tt.want.password)
			gotUser := &entity.User{}

			err = db.QueryRowContext(ctx,
				"SELECT id, login, password FROM users WHERE login = $1", tt.argument.login).
				Scan(&gotUser.ID, &gotUser.Login, &gotUser.Password)
			require.NoError(t, err)
			expectedUser.ID = gotUser.ID
			assert.Equal(t, expectedUser, gotUser)
		})
		err := clearUsersTable(ctx)
		require.NoError(t, err)
	}
}

func clearUsersTable(ctx context.Context) error {
	_, err := db.ExecContext(ctx, "DELETE FROM users")
	if err != nil {
		return fmt.Errorf("failed clear users table")
	}

	return nil
}

func TestDBStorage_HasUser(t *testing.T) {
	type fields struct {
		login    string
		password string
	}

	tests := []struct {
		name   string
		fields []fields
		args   string
		want   bool
	}{
		{
			name: "has user case",
			fields: []fields{
				{
					login:    "test",
					password: "123",
				},
				{
					login:    "test-1",
					password: "123456",
				},
			},
			args: "test",
			want: true,
		},
		{
			name:   "not has user case with empty storage",
			fields: nil,
			args:   "test",
			want:   false,
		},
		{
			name: "not has user with filled storage",
			fields: []fields{
				{
					login:    "test-123",
					password: "123456789",
				},
			},
			args: "test",
			want: false,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := db.Beginx()
			require.NoError(t, err)
			users := make([]entity.User, 0, len(tt.fields))
			for _, field := range tt.fields {
				user := entity.User{
					Login: field.login,
				}
				user.SetPassword(field.password)
				users = append(users, user)
			}
			if len(users) > 0 {
				_, err = tx.NamedExecContext(ctx, "INSERT INTO users (login, password) VALUES (:login, :password)", users)
			}
			require.NoError(t, err)
			err = tx.Commit()
			require.NoError(t, err)
			d := New(db)
			got, err := d.HasUser(ctx, tt.args)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
		err := clearUsersTable(ctx)
		require.NoError(t, err)
	}
}
