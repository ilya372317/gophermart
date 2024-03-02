package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
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

}
