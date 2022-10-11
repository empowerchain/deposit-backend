package admin

import (
	"context"

	"encore.app/commons/testutils"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsAdmin(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	require.NoError(t, InsertTestData(context.Background()))

	testTable := []struct {
		name     string
		pubKey   string
		expected bool
	}{
		{
			name:     "Is admin",
			pubKey:   testutils.AdminPubKey,
			expected: true,
		},
		{
			name:     "Is not admin",
			pubKey:   "notAdmin",
			expected: false,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			resp, err := IsAdmin(context.Background(), &IsAdminParams{PubKey: test.pubKey})
			require.NoError(t, err)
			require.Equal(t, test.expected, resp.IsAdmin)
		})
	}
}
