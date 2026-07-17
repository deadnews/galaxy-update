package galaxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_hasTimeout(t *testing.T) {
	assert.Positive(t, NewClient().http.Timeout)
}

func TestLatestCollection(t *testing.T) {
	client := startAPI(t, map[string]string{"community.general": "11.2.0"}, nil)

	t.Run("resolves highest version", func(t *testing.T) {
		got, err := client.LatestCollection(t.Context(), "community.general")
		require.NoError(t, err)
		assert.Equal(t, "11.2.0", got)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := client.LatestCollection(t.Context(), "no.collection")
		assert.Error(t, err)
	})

	t.Run("invalid name", func(t *testing.T) {
		_, err := client.LatestCollection(t.Context(), "nodot")
		assert.Error(t, err)
	})
}

func TestLatestRole(t *testing.T) {
	client := startAPI(t, nil, map[string]string{"geerlingguy.docker": "8.0.0"})

	t.Run("resolves latest version", func(t *testing.T) {
		got, err := client.LatestRole(t.Context(), "geerlingguy.docker")
		require.NoError(t, err)
		assert.Equal(t, "8.0.0", got)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := client.LatestRole(t.Context(), "no.role")
		assert.ErrorIs(t, err, errNotFound)
	})
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"8.0.0", "7.9.0", 1},
		{"7.9.0", "8.0.0", -1},
		{"1.2.3", "1.2.3", 0},
		{"1.10.0", "1.9.0", 1},
		{"2.0", "2.0.0", 0},
		{"v8.0.0", "7.9.0", 1},
		{"7.9.0", "v8.0.0", -1},
		{"v1.2.3", "1.2.3", 0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			assert.Equal(t, tt.want, compareVersions(tt.a, tt.b))
		})
	}
}
