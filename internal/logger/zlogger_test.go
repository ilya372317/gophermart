package logger

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "production test",
			env:  "prod",
		},
		{
			name: "development test",
			env:  "dev",
		},
	}
	const logText = "log is work"
	const filePath = "test-log.txt"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Log = nil
			err := Initialize(filePath)
			require.NoError(t, err)
			t.Setenv("ENV", tt.env)
			Log.Info(logText)

			logFile, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
			require.NoError(t, err)

			defer func() {
				err = logFile.Close()
				require.NoError(t, err)
			}()

			fileContent, err := io.ReadAll(logFile)
			assert.Contains(t, string(fileContent), logText)
		})
	}
	err := os.Remove(filePath)
	require.NoError(t, err)
}
