package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autobee/server/internal/auth"
)

const (
	testAccessSecret  = "test-access-secret-key-minimum-32-chars"
	testRefreshSecret = "test-refresh-secret-key-minimum-32-chars"
)

func newTestJWTService() *auth.JWTService {
	return auth.NewJWTService(
		testAccessSecret,
		testRefreshSecret,
		15*time.Minute,
		7*24*time.Hour,
	)
}

func TestGenerateTokenPair(t *testing.T) {
	svc := newTestJWTService()

	pair, err := svc.GenerateTokenPair("user123", "tenant456", auth.RoleUser)
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, pair.AccessToken, pair.RefreshToken)
}

func TestValidateAccessToken_Valid(t *testing.T) {
	svc := newTestJWTService()

	pair, err := svc.GenerateTokenPair("user123", "tenant456", auth.RoleTenantAdmin)
	require.NoError(t, err)

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.Subject)
	assert.Equal(t, "tenant456", claims.TenantID)
	assert.Equal(t, auth.RoleTenantAdmin, claims.Role)
	assert.Equal(t, auth.TokenTypeAccess, claims.Type)
	assert.NotEmpty(t, claims.JTI)
}

func TestValidateRefreshToken_Valid(t *testing.T) {
	svc := newTestJWTService()

	pair, err := svc.GenerateTokenPair("user123", "tenant456", auth.RoleUser)
	require.NoError(t, err)

	claims, err := svc.ValidateRefreshToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, auth.TokenTypeRefresh, claims.Type)
}

func TestValidateAccessToken_WrongTokenType(t *testing.T) {
	svc := newTestJWTService()

	pair, err := svc.GenerateTokenPair("user123", "tenant456", auth.RoleUser)
	require.NoError(t, err)

	// Using refresh token where access token is expected.
	// The refresh token is signed with the refresh secret, so the access token
	// validator (which uses the access secret) will see an invalid signature.
	_, err = svc.ValidateAccessToken(pair.RefreshToken)
	assert.Error(t, err)
}

func TestValidateAccessToken_InvalidSignature(t *testing.T) {
	svc := newTestJWTService()

	pair, err := svc.GenerateTokenPair("user123", "tenant456", auth.RoleUser)
	require.NoError(t, err)

	// Tamper with the token.
	tampered := pair.AccessToken + "tampered"
	_, err = svc.ValidateAccessToken(tampered)
	assert.Error(t, err)
}

func TestValidateAccessToken_Expired(t *testing.T) {
	svc := auth.NewJWTService(
		testAccessSecret,
		testRefreshSecret,
		-1*time.Second, // already expired
		7*24*time.Hour,
	)

	pair, err := svc.GenerateTokenPair("user123", "tenant456", auth.RoleUser)
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(pair.AccessToken)
	assert.Error(t, err)
}

func TestRoles(t *testing.T) {
	assert.Equal(t, auth.Role("super_admin"), auth.RoleSuperAdmin)
	assert.Equal(t, auth.Role("tenant_admin"), auth.RoleTenantAdmin)
	assert.Equal(t, auth.Role("user"), auth.RoleUser)
}
