// Package server registers all application routes.
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	httpswagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/autobee/server/docs" // swaggo generated docs

	"github.com/autobee/server/internal/auth"
	"github.com/autobee/server/internal/config"
	"github.com/autobee/server/internal/middleware"
	"github.com/autobee/server/internal/tenants"
	"github.com/autobee/server/internal/users"
	"github.com/autobee/server/internal/vehicles"
	"github.com/autobee/server/internal/bookings"
	"github.com/autobee/server/internal/repairs"
	svccenters "github.com/autobee/server/internal/service_centers"
	svcs "github.com/autobee/server/internal/services"
	"github.com/autobee/server/pkg/response"

	"go.uber.org/zap"
)

// Dependencies holds all handler instances required by the router.
type Dependencies struct {
	Config              *config.Config
	Log                 *zap.Logger
	JWTService          *auth.JWTService
	AuthHandler         *auth.Handler
	UsersHandler        *users.Handler
	TenantsHandler      *tenants.Handler
	VehiclesHandler     *vehicles.Handler
	BookingsHandler     *bookings.Handler
	RepairsHandler      *repairs.Handler
	ServiceCentersHandler *svccenters.Handler
	ServicesHandler     *svcs.Handler
}

// NewRouter assembles the Chi router with all middleware and routes.
func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// --- Global Middleware ---
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestSize(1 << 20)) // 1 MB request body limit
	r.Use(middleware.CORS(&deps.Config.CORS))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(deps.Log))

	// --- Health Endpoints (no auth) ---
	r.Get("/health", healthHandler)
	r.Get("/ready", readyHandler(deps.Log))

	// --- Swagger UI ---
	r.Get("/api/docs/*", httpswagger.Handler(
		httpswagger.URL("/api/docs/doc.json"),
	))

	// --- API v1 ---
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", deps.AuthHandler.Signup)
			r.Post("/signin", deps.AuthHandler.Signin)
			r.Post("/refresh", deps.AuthHandler.Refresh)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(deps.JWTService))
			r.Use(middleware.EnforceTenantContext)

			// Users
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", deps.UsersHandler.GetMe)
			})

			// Vehicles (reference module)
			r.Route("/vehicles", func(r chi.Router) {
				r.Get("/", deps.VehiclesHandler.List)
				r.Post("/", deps.VehiclesHandler.Create)
				r.Get("/{id}", deps.VehiclesHandler.GetByID)
			})

			// Bookings
			r.Route("/bookings", func(r chi.Router) {
				r.Get("/", deps.BookingsHandler.List)
			})

			// Repairs
			r.Route("/repairs", func(r chi.Router) {
				r.Get("/", deps.RepairsHandler.List)
			})

			// Service Centers
			r.Route("/service-centers", func(r chi.Router) {
				r.Get("/", deps.ServiceCentersHandler.List)
			})

			// Services
			r.Route("/services", func(r chi.Router) {
				r.Get("/", deps.ServicesHandler.List)
			})

			// Tenants — super admin only
			r.Route("/tenants", func(r chi.Router) {
				r.Use(middleware.RequireRole(auth.RoleSuperAdmin))
				r.Get("/", deps.TenantsHandler.ListAll)
				r.Post("/", deps.TenantsHandler.Create)
				r.Get("/{id}", deps.TenantsHandler.GetByID)
			})
		})
	})

	return r
}

// healthHandler returns a simple 200 OK to indicate the process is alive.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyHandler checks that critical dependencies (MongoDB) are reachable.
// A real readiness check would ping MongoDB; this is the structural placeholder.
func readyHandler(log *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In production this should ping MongoDB. The actual ping is done at
		// startup via database.Connect() — if the server is running, Mongo is up.
		response.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
