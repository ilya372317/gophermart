package luhnalgo

import "testing"

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		argument int
		want     bool
	}{
		{
			name:     "positive case",
			argument: 4324802833166747,
			want:     true,
		},
		{
			name:     "invalid case",
			argument: 123,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.argument); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
