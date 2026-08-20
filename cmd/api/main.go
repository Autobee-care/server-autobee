// Package main is the entry point for the autobee-server API.
//
//	@title			Autobee Server API
//	@version		1.0
//	@description	Multi-tenant vehicle service management platform API
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	Autobee Support
//	@contact.email	support@autobee.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and the JWT access token
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
	"github.com/autobee/server/pkg/logger"
)

func main() {
	// --- Configuration ---
	cfg, err := config.Load()
	if err != nil {
		// Use stdlib log here because Zap isn't initialized yet.
		panic("config error: " + err.Error())
	}

	// --- Logger ---
	log, err := logger.New(cfg.LogLevel, cfg.App.Env)
	if err != nil {
		panic("logger init error: " + err.Error())
	}
	defer log.Sync() //nolint:errcheck

	log.Info("starting", zap.String("app", cfg.App.Name), zap.String("env", cfg.App.Env))

	// --- MongoDB ---
	ctx := context.Background()
	dbClient, err := database.Connect(ctx, cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		log.Fatal("failed to connect to MongoDB", zap.Error(err))
	}
	log.Info("connected to MongoDB", zap.String("database", cfg.Mongo.Database))

	// --- Indexes ---
	if err := database.EnsureIndexes(ctx, dbClient.Database); err != nil {
		log.Fatal("failed to create indexes", zap.Error(err))
	}
	log.Info("indexes ensured")

	// --- Services ---
	jwtSvc := auth.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiration,
		cfg.JWT.RefreshExpiration,
	)
	otpSvc := auth.NewNoOpOTPService(log)

	// Auth
	authRepo := auth.NewRepository(dbClient.Collection("users"))
	authSvc := auth.NewService(authRepo, jwtSvc, otpSvc, log)
	authHandler := auth.NewHandler(authSvc, log)

	// Users
	usersRepo := users.NewRepository(dbClient.Collection("users"))
	usersSvc := users.NewService(usersRepo, log)
	usersHandler := users.NewHandler(usersSvc, log)

	// Tenants
	tenantsRepo := tenants.NewRepository(dbClient.Collection("tenants"))
	tenantsSvc := tenants.NewService(tenantsRepo, log)
	tenantsHandler := tenants.NewHandler(tenantsSvc, log)

	// Vehicles (reference module)
	vehiclesRepo := vehicles.NewRepository(dbClient.Collection("vehicles"))
	vehiclesSvc := vehicles.NewService(vehiclesRepo, log)
	vehiclesHandler := vehicles.NewHandler(vehiclesSvc, log)

	// Stub modules
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

	// --- Router ---
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

	// --- HTTP Server ---
	srv := server.New(cfg.App.Port, router, log)

	// Start server in background goroutine.
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received")

	if err := srv.Shutdown(); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}

	if err := dbClient.Disconnect(ctx); err != nil {
		log.Error("mongo disconnect error", zap.Error(err))
	}

	log.Info("server stopped gracefully")
}
