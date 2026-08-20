package users_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/autobee/server/internal/users"
)

func TestGetMe_InvalidUserID(t *testing.T) {
	log := zap.NewNop()
	// Use nil repo — invalid ID format should fail before any DB call.
	repo := (*users.Repository)(nil)
	svc := users.NewService(repo, log)

	_, err := svc.GetMe(context.Background(), "not-a-valid-object-id")
	assert.ErrorIs(t, err, users.ErrInvalidUserID)
}
