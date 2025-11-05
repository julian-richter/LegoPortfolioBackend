package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"LegoManagerAPI/internal/api/handlers"
	health2 "LegoManagerAPI/internal/api/handlers/health"
	checks2 "LegoManagerAPI/internal/api/handlers/health/checks"
	"LegoManagerAPI/internal/cache"
	"LegoManagerAPI/internal/config"
	"LegoManagerAPI/internal/database"
	"LegoManagerAPI/internal/repos"
)

type Server struct {
	httpServer    *http.Server
	cfg           *config.Config
	HealthService *health2.Service
}

func NewServer(cfg *config.Config, db *database.PostgresDB, redisClient *cache.RedisClient) *Server {
	// Health checks
	healthCheckers := []health2.Checker{
		checks2.NewPostgresCheck(db),
		checks2.NewRedisCheck(redisClient),
		checks2.NewApplicationCheck(),
	}
	healthService := health2.NewService(cfg.App.Environment, healthCheckers...)

	// Initialize repositories
	userRepo := repos.NewUserRepository(db.Pool)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(healthService)
	userHandler := handlers.NewUserHandler(userRepo)

	// Setup router
	router := http.NewServeMux()

	// Register routes
	router.HandleFunc("/", handleRoot)
	router.HandleFunc("/health", healthHandler.Handle)

	// User routes
	router.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Check if it's a search
			if r.URL.Query().Get("q") != "" {
				userHandler.SearchUsers(w, r)
			} else {
				userHandler.ListUsers(w, r)
			}
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	router.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a password update
		if strings.HasSuffix(r.URL.Path, "/password") {
			if r.Method == http.MethodPost {
				userHandler.UpdatePassword(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Regular user CRUD
		switch r.Method {
		case http.MethodGet:
			userHandler.GetUser(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer:    server,
		cfg:           cfg,
		HealthService: healthService,
	}
}

func (s *Server) Start() error {
	log.Info("Starting HTTP server", "port", s.cfg.App.Port)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info("HTTP server stopped")
	return nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World!"))
}
