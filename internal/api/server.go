package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/log"

	"LegoManagerAPI/internal/api/handlers"
	"LegoManagerAPI/internal/cache"
	"LegoManagerAPI/internal/config"
	"LegoManagerAPI/internal/database"
	"LegoManagerAPI/internal/health"
	"LegoManagerAPI/internal/health/checks"
)

type Server struct {
	httpServer    *http.Server
	cfg           *config.Config
	HealthService *health.Service
}

func NewServer(cfg *config.Config, db *database.PostgresDB, redisClient *cache.RedisClient) *Server {
	healthCheckers := []health.Checker{
		checks.NewPostgresCheck(db),
		checks.NewRedisCheck(redisClient),
		checks.NewApplicationCheck(),
	}

	healthService := health.NewService(cfg.App.Environment, healthCheckers...)

	router := http.NewServeMux()

	// Register routes
	router.HandleFunc("/", handleRoot)

	// Add health endpoint
	healthHandler := handlers.NewHealthHandler(healthService)
	router.HandleFunc("/health", healthHandler.Handle)

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
	log.Info("Starting HTTP server on port ", s.cfg.App.Port)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server faile: %w", err)
		return nil
	}

	log.Info("Server stopped")
	return nil
}

// Handlers
func handleRoot(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/plain")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Hello World!"))
}
