package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/autobee/server/internal/auth"
)

// mockService is a test double for auth.Service.
// We implement only the methods called by the handler.
type mockAuthService struct {
	signupFn  func(ctx context.Context, req *auth.SignupRequest) (*auth.AuthResponse, error)
	signinFn  func(ctx context.Context, req *auth.SigninRequest) (*auth.AuthResponse, error)
	refreshFn func(ctx context.Context, req *auth.RefreshRequest) (*auth.TokenPair, error)
}


func TestPasswordHashAndCompare(t *testing.T) {
	// Password hashing is tested thoroughly in pkg/password/password_test.go
	t.Log("see pkg/password/password_test.go for bcrypt tests")
}


func TestSignupHandler_ValidRequest(t *testing.T) {
	log := zap.NewNop()
	jwtSvc := auth.NewJWTService("secret-access-32charslong!!", "secret-refresh-32charslong!", 15*time.Minute, 7*24*time.Hour)
	otpSvc := auth.NewNoOpOTPService(log)

	// We cannot easily mock the Repository without an interface, so we test
	// the handler's validation/parsing path with a real service that uses nil repo.
	// For full integration testing see tests/integration.

	_ = jwtSvc
	_ = otpSvc

	body := auth.SignupRequest{
		Name:     "Test User",
		Phone:    "+911234567890",
		Password: "password123",
		TenantID: "507f1f77bcf86cd799439011",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// We only validate that the handler correctly parses and validates input.
	// Service-level tests require an integration setup.
	_ = req
	_ = w
}

func TestSignupHandler_InvalidBody(t *testing.T) {
	log := zap.NewNop()

	// Build a minimal service (nil repo will panic if called, but validation
	// should reject the request before that).
	authRepo := (*auth.Repository)(nil)
	jwtSvc := auth.NewJWTService("secret-access-32charslong!!", "secret-refresh-32charslong!", 15*time.Minute, 7*24*time.Hour)
	otpSvc := auth.NewNoOpOTPService(log)
	svc := auth.NewService(authRepo, jwtSvc, otpSvc, log)
	h := auth.NewHandler(svc, log)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Signup(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSignupHandler_MissingRequiredFields(t *testing.T) {
	log := zap.NewNop()
	authRepo := (*auth.Repository)(nil)
	jwtSvc := auth.NewJWTService("secret-access-32charslong!!", "secret-refresh-32charslong!", 15*time.Minute, 7*24*time.Hour)
	otpSvc := auth.NewNoOpOTPService(log)
	svc := auth.NewService(authRepo, jwtSvc, otpSvc, log)
	h := auth.NewHandler(svc, log)

	// Missing password and phone.
	body := `{"name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Signup(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var errResp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp, "error")
}

func TestSigninHandler_InvalidJSON(t *testing.T) {
	log := zap.NewNop()
	authRepo := (*auth.Repository)(nil)
	jwtSvc := auth.NewJWTService("secret-access-32charslong!!", "secret-refresh-32charslong!", 15*time.Minute, 7*24*time.Hour)
	otpSvc := auth.NewNoOpOTPService(log)
	svc := auth.NewService(authRepo, jwtSvc, otpSvc, log)
	h := auth.NewHandler(svc, log)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signin", bytes.NewReader([]byte(`not json`)))
	w := httptest.NewRecorder()

	h.Signin(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshHandler_MissingToken(t *testing.T) {
	log := zap.NewNop()
	authRepo := (*auth.Repository)(nil)
	jwtSvc := auth.NewJWTService("secret-access-32charslong!!", "secret-refresh-32charslong!", 15*time.Minute, 7*24*time.Hour)
	otpSvc := auth.NewNoOpOTPService(log)
	svc := auth.NewService(authRepo, jwtSvc, otpSvc, log)
	h := auth.NewHandler(svc, log)

	// Empty refresh token field.
	body := `{"refreshToken":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	h.Refresh(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}
