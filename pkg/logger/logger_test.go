package logger_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/htrandev/gophermart/pkg/logger"
)

func TestNewLogger(t *testing.T) {
	testCases := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "valid",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "valid",
			level:   "test",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		_, err := logger.NewZapLogger(tc.level)
		if tc.wantErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
	}
}
