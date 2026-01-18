package web

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	db     *sql.DB
	router *http.ServeMux
	port   int
}

func NewServer(db *sql.DB, port int) *Server {
	s := &Server{
		db:     db,
		router: http.NewServeMux(),
		port:   port,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Static files
	staticFS, _ := fs.Sub(staticFiles, "static")
	s.router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Pages
	s.router.HandleFunc("GET /", s.handleDashboard)
	s.router.HandleFunc("GET /sessions", s.handleSessions)
	s.router.HandleFunc("GET /sessions/{id}", s.handleSessionDetail)
	s.router.HandleFunc("GET /experiments", s.handleExperiments)
	s.router.HandleFunc("GET /experiments/compare", s.handleExperimentCompare)
	s.router.HandleFunc("GET /experiments/{id}", s.handleExperimentDetail)
	s.router.HandleFunc("GET /settings", s.handleSettings)

	// API endpoints (for HTMX)
	s.router.HandleFunc("GET /api/stats", s.handleAPIStats)
	s.router.HandleFunc("GET /api/charts/tokens", s.handleAPIChartTokens)
	s.router.HandleFunc("GET /api/charts/cost", s.handleAPIChartCost)
	s.router.HandleFunc("POST /api/experiments", s.handleAPICreateExperiment)
	s.router.HandleFunc("POST /api/experiments/{id}/activate", s.handleAPIActivateExperiment)
	s.router.HandleFunc("POST /api/experiments/{id}/deactivate", s.handleAPIDeactivateExperiment)
	s.router.HandleFunc("DELETE /api/experiments/{id}", s.handleAPIDeleteExperiment)
}

func (s *Server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Starting server at http://localhost:%d\n", s.port)

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}
