package password_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autobee/server/pkg/password"
)

func TestHash_ReturnsNonEmptyHash(t *testing.T) {
	hash, err := password.Hash("mysecretpassword")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$"),
		"expected bcrypt hash prefix, got: %s", hash)
}

func TestHash_DifferentHashesForSamePassword(t *testing.T) {
	h1, err := password.Hash("samepassword")
	require.NoError(t, err)

	h2, err := password.Hash("samepassword")
	require.NoError(t, err)

	// bcrypt uses random salt, so hashes must differ.
	assert.NotEqual(t, h1, h2)
}

func TestCompare_CorrectPassword(t *testing.T) {
	hash, err := password.Hash("correctpassword")
	require.NoError(t, err)

	assert.True(t, password.Compare(hash, "correctpassword"))
}

func TestCompare_WrongPassword(t *testing.T) {
	hash, err := password.Hash("correctpassword")
	require.NoError(t, err)

	assert.False(t, password.Compare(hash, "wrongpassword"))
}

func TestCompare_EmptyPassword(t *testing.T) {
	hash, err := password.Hash("correctpassword")
	require.NoError(t, err)

	assert.False(t, password.Compare(hash, ""))
}
