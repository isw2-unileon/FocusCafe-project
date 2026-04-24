// Package main is the entry point for the backend server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/config"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/supabase"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	ctx := context.Background()
	cfg := config.Load()

	database.InitDB(cfg)

	gin.SetMode(cfg.GinMode)

	// Create jwt adapter
	adapterJWT, err := supabase.NewJWTAdapter(cfg.SupabaseURL)
	if err != nil {
		logger.Error("failed to create jwt adapter", "error", err)
		os.Exit(1)
	}

	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.CORSAllowOrigin},
		AllowMethods: []string{"POST", "GET", "PUT", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	r.Use(gin.Logger(), gin.Recovery())

	h := &handlers.Handler{
		SupabaseURL: cfg.SupabaseURL,
		SupabaseKey: cfg.SupabaseKey,
		Auth:        adapterJWT,
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes
	api := r.Group("/api")
	api.POST("/login", h.Login)
	api.POST("/register", h.Register)
	api.GET("/auth/google", h.GoogleAuth)
	api.POST("/auth/sync", h.SyncUser)

	// Protected route
	protected := api.Group("/")
	protected.Use(handlers.Auth(adapterJWT))
	protected.GET("/users/me", h.GetUserProfile)
	protected.PUT("/users/me", h.UpdateUserProfile)
	protected.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from the API"})
	})

	protected.POST("/study/start", handlers.StartStudySessionHandler)
	protected.POST("/study/generate-quiz/:session_id", handlers.CreateQuizFromSession)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	logger.Info("server stopped")

	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		_ = os.Mkdir("uploads", 0o750)
	}
}
