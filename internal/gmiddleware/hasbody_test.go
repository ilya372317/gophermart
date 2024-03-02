package gmiddleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldHasBody(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "empty body case",
			body: "",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "not empty body case",
			body: "some body",
			want: want{
				code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/some-route", bytes.NewReader([]byte(tt.body)))
			writer := httptest.NewRecorder()
			handler := ShouldHasBody(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))
			handler.ServeHTTP(writer, request)
			res := writer.Result()
			err := res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
