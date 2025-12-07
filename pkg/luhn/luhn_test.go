package luhn

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	testCase := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "valid",
			value: "5555555555554444",
			valid: true,
		},
		{
			name:  "invalid",
			value: "411111111111111",
			valid: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			valid := Check(tc.value)
			require.Equal(t, tc.valid, valid)
		})
	}
}
