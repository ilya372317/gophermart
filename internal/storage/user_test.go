package storage

import (
	"context"
	"database/sql"
	"errors"
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

type userFields struct {
	login    string
	password string
	balance  int
}

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

func TestDBStorage_GetUserByLogin(t *testing.T) {
	tests := []struct {
		name     string
		fields   []userFields
		argument string
		wantErr  bool
		want     *entity.User
	}{
		{
			name: "success case",
			fields: []userFields{
				{
					login:    "login",
					password: "pass",
				},
				{
					login:    "second user",
					password: "second pass",
				},
			},
			argument: "login",
			wantErr:  false,
			want: &entity.User{
				Login:    "login",
				Password: "pass",
			},
		},
		{
			name:     "not found case",
			fields:   nil,
			argument: "test",
			wantErr:  true,
			want:     nil,
		},
	}
	for _, tt := range tests {
		ctx := context.Background()
		t.Run(tt.name, func(t *testing.T) {
			fillUsers(ctx, t, tt.fields)
			repo := New(db)
			got, err := repo.GetUserByLogin(ctx, tt.argument)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			if tt.want != nil {
				tt.want.SetPassword(tt.want.Password)
				tt.want.CreatedAT = got.CreatedAT
				tt.want.UpdatedAT = got.UpdatedAT
			}
			assert.Equal(t, tt.want.Login, got.Login)
			assert.Equal(t, tt.want.Password, got.Password)
		})
		clearUsersTable(ctx, t)
	}
}

func TestDBStorage_GetUserByID(t *testing.T) {
	tests := []struct {
		name    string
		fields  []userFields
		want    *entity.User
		wantErr bool
	}{
		{
			name: "success filled storage case",
			fields: []userFields{
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
			want: &entity.User{
				ID:       2,
				Login:    "543",
				Password: "password-543",
			},
			wantErr: false,
		},
		{
			name:    "user not found case",
			fields:  nil,
			want:    nil,
			wantErr: true,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fillUsers(ctx, t, tt.fields)
			repo := New(db)
			var lastID sql.NullInt64
			var lastIDValue uint
			err := db.Get(&lastID, "SELECT MAX(id) FROM users")
			if lastID.Valid {
				lastIDValue = uint(lastID.Int64)
			}
			require.NoError(t, err)
			got, err := repo.GetUserByID(ctx, lastIDValue)
			if errors.Is(err, sql.ErrNoRows) {
				if tt.wantErr {
					require.Error(t, err)
					return
				} else {
					require.NoError(t, err)
				}
			} else {
				require.NoError(t, err)
			}
			if tt.want != nil {
				tt.want.SetPassword(tt.want.Password)
				tt.want.CreatedAT = got.CreatedAT
				tt.want.UpdatedAT = got.UpdatedAT
			}
			assert.Equal(t, tt.want.Login, got.Login)
			assert.Equal(t, tt.want.Password, got.Password)
		})
		clearUsersTable(ctx, t)
	}
}

func TestDBStorage_SaveUser(t *testing.T) {
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
		fields   []userFields
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
			fields: []userFields{
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
			fields: []userFields{
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
			assert.Equal(t, expectedUser.Login, gotUser.Login)
			assert.Equal(t, expectedUser.Password, gotUser.Password)
		})
		clearUsersTable(ctx, t)
	}
}

func clearUsersTable(ctx context.Context, t *testing.T) {
	t.Helper()
	_, err := db.ExecContext(ctx, "DELETE FROM users")
	require.NoError(t, err)
}

func TestDBStorage_HasUser(t *testing.T) {
	tests := []struct {
		name   string
		fields []userFields
		args   string
		want   bool
	}{
		{
			name: "has user case",
			fields: []userFields{
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
			fields: []userFields{
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
		clearUsersTable(ctx, t)
	}
}

func fillUsers(ctx context.Context, t *testing.T, fields []userFields) {
	t.Helper()
	users := make([]entity.User, 0, len(fields))
	for _, f := range fields {
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
}

func TestDBStorage_UpdateUserBalanceByID(t *testing.T) {
	tests := []struct {
		field        userFields
		name         string
		argument     float64
		wantErr      bool
		userIDExists bool
	}{
		{
			name: "success case",
			field: userFields{
				login:    "123",
				password: "123",
				balance:  10,
			},
			wantErr:      false,
			argument:     20,
			userIDExists: true,
		},
		{
			name: "user id not exists case",
			field: userFields{
				login:    "123",
				password: "123",
				balance:  10,
			},
			wantErr:      true,
			argument:     20,
			userIDExists: false,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearUsersTable(ctx, t)
			fields := make([]userFields, 0)
			fields = append(fields, tt.field)
			fillUsers(ctx, t, fields)
			repo := New(db)
			userID := getLastUserID(ctx, t)
			if !tt.userIDExists {
				userID = 0
			}
			err := repo.UpdateUserBalanceByID(ctx, userID, tt.argument)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			got := getLastUser(ctx, t)
			assert.Equal(t, tt.argument, got.Balance)
		})
	}
}

func getLastUser(ctx context.Context, t *testing.T) *entity.User {
	t.Helper()
	user := &entity.User{}
	err := db.GetContext(ctx, user, "SELECT * FROM users WHERE id = (SELECT MAX(id) FROM users)")
	require.NoError(t, err)
	return user
}

func getLastUserID(ctx context.Context, t *testing.T) uint {
	t.Helper()
	var userID uint
	err := db.GetContext(ctx, &userID, "SELECT MAX(id) FROM users")
	require.NoError(t, err)
	return userID
}
