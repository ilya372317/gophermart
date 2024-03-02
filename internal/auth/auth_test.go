package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ilya372317/gophermart/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GenerateJWTToken(t *testing.T) {
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
		{
			name: "invalid user id case",
			fields: fields{
				Login:    "123",
				Password: "123",
				ID:       0,
			},
			args: args{
				secretKey: "secret-key",
				expTime:   time.Second,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &entity.User{
				Login:    tt.fields.Login,
				Password: tt.fields.Password,
				ID:       tt.fields.ID,
			}
			sut := New(tt.args.secretKey, tt.args.expTime)
			got, err := sut.GenerateJWTToken(user)
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
