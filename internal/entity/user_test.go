package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_CheckPassword(t *testing.T) {
	type fields struct {
		ID       uint
		Login    string
		password string
	}
	type args struct {
		pass string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "correct Password case",
			fields: fields{
				ID:       1,
				Login:    "test",
				password: "123",
			},
			args: args{
				pass: "123",
			},
			want: true,
		},
		{
			name: "invalid Password case",
			fields: fields{
				ID:       1,
				Login:    "test",
				password: "1",
			},
			args: args{
				pass: "123",
			},
			want: false,
		},
		{
			name: "empty argument case",
			fields: fields{
				ID:       1,
				Login:    "test",
				password: "123",
			},
			args: args{
				pass: "",
			},
			want: false,
		},
		{
			name: "empty argument and Password case",
			fields: fields{
				ID:       1,
				Login:    "test",
				password: "",
			},
			args: args{
				pass: "",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:    tt.fields.ID,
				Login: tt.fields.Login,
			}
			u.SetPassword(tt.fields.password)
			if got := u.CheckPassword(tt.args.pass); got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_GetPasswordHash(t *testing.T) {
	type fields struct {
		ID       uint
		Login    string
		password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty hash case",
			fields: fields{
				ID:       1,
				Login:    "test",
				password: "",
			},
			want: "",
		},
		{
			name: "success case",
			fields: fields{
				ID:       1,
				Login:    "test",
				password: "123",
			},
			want: "123",
		},
		{
			name: "long hash case",
			fields: fields{
				ID:       1,
				Login:    "test",
				password: "12345678910qwertyuioop[asdfghklxcv,bkgfkskzkakskdkfkgjbjcjzkkasdkkdgfjfgdjjcvx",
			},
			want: "12345678910qwertyuioop[asdfghklxcv,bkgfkskzkakskdkfkgjbjcjzkkasdkkdgfjfgdjjcvx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:    tt.fields.ID,
				Login: tt.fields.Login,
			}
			u.SetPassword(tt.fields.password)
			hash := sha256.Sum256([]byte(tt.want))
			stringHash := hex.EncodeToString(hash[:])
			if got := u.GetPasswordHash(); got != stringHash {
				t.Errorf("GetPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_GenerateJWTToken(t *testing.T) {
	type fields struct {
		Login    string
		Password string
		ID       uint
	}
	type args struct {
		secretKey string
		expTime   time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			fields: fields{
				Login:    "123",
				Password: "password",
				ID:       10,
			},
			args: args{
				secretKey: "secret-key",
				expTime:   time.Hour * 1,
			},
			wantErr: false,
		},
		{
			name: "negative empty secret case",
			fields: fields{
				Login:    "123",
				Password: "password",
				ID:       10,
			},
			args: args{
				secretKey: "",
				expTime:   time.Hour,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				Login:    tt.fields.Login,
				Password: tt.fields.Password,
				ID:       tt.fields.ID,
			}
			got, err := u.GenerateJWTToken(tt.args.secretKey, tt.args.expTime)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(got, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("invalid signing method: %w", err)
				}
				return []byte(tt.args.secretKey), nil
			})

			assert.True(t, token.Valid)
			assert.Equal(t, tt.fields.ID, claims.UserID)
		})
	}
}
