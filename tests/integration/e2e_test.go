//go:build integration

// Package integration provides end-to-end integration tests using Testcontainers.
// MongoDB is started automatically — no manual setup required.
// Run with: make test-integration  OR  go test -tags=integration ./tests/integration/...
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.uber.org/zap"

	"github.com/autobee/server/internal/auth"
	"github.com/autobee/server/internal/bookings"
	"github.com/autobee/server/internal/config"
	"github.com/autobee/server/internal/database"
	"github.com/autobee/server/internal/repairs"
	svccenters "github.com/autobee/server/internal/service_centers"
	svcs "github.com/autobee/server/internal/services"
	"github.com/autobee/server/internal/server"
	"github.com/autobee/server/internal/tenants"
	"github.com/autobee/server/internal/users"
	"github.com/autobee/server/internal/vehicles"
)

// testApp bundles the test HTTP server and shared test state.
type testApp struct {
	srv      *httptest.Server
	tenantID string
	dbClient *database.Client
}

// setupTestApp starts MongoDB via Testcontainers and wires the full application.
func setupTestApp(t *testing.T) *testApp {
	t.Helper()
	ctx := context.Background()

	// Start MongoDB container.
	mongoContainer, err := mongodb.Run(ctx, "mongo:7")
	require.NoError(t, err, "failed to start MongoDB container")

	t.Cleanup(func() {
		if err := mongoContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate MongoDB container: %v", err)
		}
	})

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	dbClient, err := database.Connect(ctx, mongoURI, "autobee_test")
	require.NoError(t, err, "failed to connect to test MongoDB")

	t.Cleanup(func() {
		_ = dbClient.Disconnect(ctx)
	})

	require.NoError(t, database.EnsureIndexes(ctx, dbClient.Database))

	// Create a test tenant first.
	tenantsRepo := tenants.NewRepository(dbClient.Collection("tenants"))
	testTenant := &tenants.Tenant{
		ID:        bson.NewObjectID(),
		Name:      "Test Tenant",
		Status:    tenants.TenantStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	createdTenant, err := tenantsRepo.Create(ctx, testTenant)
	require.NoError(t, err)

	log := zap.NewNop()

	cfg := &config.Config{
		App:      config.AppConfig{Env: "test", Name: "autobee-test", Port: "0"},
		LogLevel: "error",
		CORS: config.CORSConfig{
			AllowedOrigins:   []string{"http://localhost:3000"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
		},
	}

	jwtSvc := auth.NewJWTService(
		"test-access-secret-minimum-32chars!!",
		"test-refresh-secret-minimum-32chars!",
		15*time.Minute,
		7*24*time.Hour,
	)
	otpSvc := auth.NewNoOpOTPService(log)

	authRepo := auth.NewRepository(dbClient.Collection("users"))
	authSvc := auth.NewService(authRepo, jwtSvc, otpSvc, log)
	authHandler := auth.NewHandler(authSvc, log)

	usersRepo := users.NewRepository(dbClient.Collection("users"))
	usersSvc := users.NewService(usersRepo, log)
	usersHandler := users.NewHandler(usersSvc, log)

	tenantsSvc := tenants.NewService(tenantsRepo, log)
	tenantsHandler := tenants.NewHandler(tenantsSvc, log)

	vehiclesRepo := vehicles.NewRepository(dbClient.Collection("vehicles"))
	vehiclesSvc := vehicles.NewService(vehiclesRepo, log)
	vehiclesHandler := vehicles.NewHandler(vehiclesSvc, log)

	bookingsRepo := bookings.NewRepository(dbClient.Collection("bookings"))
	bookingsSvc := bookings.NewService(bookingsRepo, log)
	bookingsHandler := bookings.NewHandler(bookingsSvc, log)

	repairsRepo := repairs.NewRepository(dbClient.Collection("repairs"))
	repairsSvc := repairs.NewService(repairsRepo, log)
	repairsHandler := repairs.NewHandler(repairsSvc, log)

	scRepo := svccenters.NewRepository(dbClient.Collection("service_centers"))
	scSvc := svccenters.NewService(scRepo, log)
	scHandler := svccenters.NewHandler(scSvc, log)

	svcRepo := svcs.NewRepository(dbClient.Collection("services"))
	svcSvc := svcs.NewService(svcRepo, log)
	svcHandler := svcs.NewHandler(svcSvc, log)

	router := server.NewRouter(server.Dependencies{
		Config:                cfg,
		Log:                   log,
		JWTService:            jwtSvc,
		AuthHandler:           authHandler,
		UsersHandler:          usersHandler,
		TenantsHandler:        tenantsHandler,
		VehiclesHandler:       vehiclesHandler,
		BookingsHandler:       bookingsHandler,
		RepairsHandler:        repairsHandler,
		ServiceCentersHandler: scHandler,
		ServicesHandler:       svcHandler,
	})

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	return &testApp{
		srv:      srv,
		tenantID: createdTenant.ID.Hex(),
		dbClient: dbClient,
	}
}

// post sends a POST request and returns the response.
func (a *testApp) post(t *testing.T, path string, body any, token string) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, a.srv.URL+path, bytes.NewReader(b))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// get sends a GET request and returns the response.
func (a *testApp) get(t *testing.T, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, a.srv.URL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// decodeBody JSON-decodes the response body into target.
func decodeBody(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
}

// TestE2E_SignupSigninVehicle tests the full happy path:
// Signup → Signin → Create vehicle → Get vehicle.
func TestE2E_SignupSigninVehicle(t *testing.T) {
	app := setupTestApp(t)

	phone := fmt.Sprintf("+91%d", time.Now().UnixNano()%9000000000+1000000000)

	// 1. Signup
	signupResp := app.post(t, "/api/v1/auth/signup", map[string]any{
		"name":     "Integration Test User",
		"phone":    phone,
		"password": "testpassword123",
		"tenantId": app.tenantID,
	}, "")

	assert.Equal(t, http.StatusCreated, signupResp.StatusCode)

	var signupBody struct {
		Data struct {
			User   map[string]any `json:"user"`
			Tokens struct {
				AccessToken  string `json:"accessToken"`
				RefreshToken string `json:"refreshToken"`
			} `json:"tokens"`
		} `json:"data"`
	}
	decodeBody(t, signupResp, &signupBody)
	accessToken := signupBody.Data.Tokens.AccessToken
	refreshToken := signupBody.Data.Tokens.RefreshToken
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// 2. Signin
	signinResp := app.post(t, "/api/v1/auth/signin", map[string]any{
		"phone":    phone,
		"password": "testpassword123",
		"tenantId": app.tenantID,
	}, "")
	assert.Equal(t, http.StatusOK, signinResp.StatusCode)

	var signinBody struct {
		Data struct {
			Tokens struct {
				AccessToken string `json:"accessToken"`
			} `json:"tokens"`
		} `json:"data"`
	}
	decodeBody(t, signinResp, &signinBody)
	signinToken := signinBody.Data.Tokens.AccessToken
	assert.NotEmpty(t, signinToken)

	// 3. Get /users/me
	meResp := app.get(t, "/api/v1/users/me", signinToken)
	assert.Equal(t, http.StatusOK, meResp.StatusCode)

	// 4. Create Vehicle
	createVehicleResp := app.post(t, "/api/v1/vehicles", map[string]any{
		"registrationNumber": "KA01AB1234",
		"make":               "Toyota",
		"model":              "Camry",
		"year":               2022,
		"fuelType":           "petrol",
	}, signinToken)
	assert.Equal(t, http.StatusCreated, createVehicleResp.StatusCode)

	var vehicleBody struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	decodeBody(t, createVehicleResp, &vehicleBody)
	vehicleID := vehicleBody.Data.ID
	assert.NotEmpty(t, vehicleID)

	// 5. Get Vehicle by ID
	getVehicleResp := app.get(t, "/api/v1/vehicles/"+vehicleID, signinToken)
	assert.Equal(t, http.StatusOK, getVehicleResp.StatusCode)

	// 6. Refresh token
	refreshResp := app.post(t, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": refreshToken,
	}, "")
	assert.Equal(t, http.StatusOK, refreshResp.StatusCode)

	// 7. List vehicles
	listResp := app.get(t, "/api/v1/vehicles", signinToken)
	assert.Equal(t, http.StatusOK, listResp.StatusCode)
}

// TestE2E_CrossTenantAccessRejected verifies that a user from Tenant A
// cannot access a vehicle from Tenant B.
func TestE2E_CrossTenantAccessRejected(t *testing.T) {
	app := setupTestApp(t)
	ctx := context.Background()

	// Create a second tenant.
	tenantsRepo := tenants.NewRepository(app.dbClient.Collection("tenants"))
	tenant2, err := tenantsRepo.Create(ctx, &tenants.Tenant{Name: "Tenant B"})
	require.NoError(t, err)

	phoneA := fmt.Sprintf("+91%d", time.Now().UnixNano()%9000000000+1000000000)
	phoneB := fmt.Sprintf("+92%d", time.Now().UnixNano()%9000000000+1000000000)

	// Sign up user in Tenant A.
	respA := app.post(t, "/api/v1/auth/signup", map[string]any{
		"name":     "User A",
		"phone":    phoneA,
		"password": "password123",
		"tenantId": app.tenantID,
	}, "")
	require.Equal(t, http.StatusCreated, respA.StatusCode)
	var bodyA struct {
		Data struct{ Tokens struct{ AccessToken string } }
	}
	decodeBody(t, respA, &bodyA)
	tokenA := bodyA.Data.Tokens.AccessToken

	// Sign up user in Tenant B.
	respB := app.post(t, "/api/v1/auth/signup", map[string]any{
		"name":     "User B",
		"phone":    phoneB,
		"password": "password123",
		"tenantId": tenant2.ID.Hex(),
	}, "")
	require.Equal(t, http.StatusCreated, respB.StatusCode)
	var bodyB struct {
		Data struct{ Tokens struct{ AccessToken string } }
	}
	decodeBody(t, respB, &bodyB)
	tokenB := bodyB.Data.Tokens.AccessToken

	// User B creates a vehicle.
	createResp := app.post(t, "/api/v1/vehicles", map[string]any{
		"registrationNumber": "MH01XY9999",
		"make":               "Honda",
		"model":              "City",
		"year":               2023,
		"fuelType":           "diesel",
	}, tokenB)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	var createBody struct {
		Data struct{ ID string }
	}
	decodeBody(t, createResp, &createBody)
	vehicleID := createBody.Data.ID

	// User A tries to access User B's vehicle — must be rejected with 403 or 404.
	getResp := app.get(t, "/api/v1/vehicles/"+vehicleID, tokenA)
	assert.True(t,
		getResp.StatusCode == http.StatusForbidden || getResp.StatusCode == http.StatusNotFound,
		"expected 403 or 404, got %d", getResp.StatusCode)
}

// TestHealth checks the /health endpoint.
func TestHealth(t *testing.T) {
	app := setupTestApp(t)
	resp := app.get(t, "/health", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestReady checks the /ready endpoint.
func TestReady(t *testing.T) {
	app := setupTestApp(t)
	resp := app.get(t, "/ready", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestProtectedRoute_NoToken verifies 401 for unauthenticated access.
func TestProtectedRoute_NoToken(t *testing.T) {
	app := setupTestApp(t)
	resp := app.get(t, "/api/v1/users/me", "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
