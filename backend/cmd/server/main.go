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
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/supabase"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	ctx := context.Background()
	cfg := config.Load()

	database.InitDB(cfg)

	adapterJWT, err := initJWTAdapter(cfg.SupabaseURL)
	if err != nil {
		logger.Error("failed to initialize", "error", err)
		os.Exit(1)
	}

	userService, userOrderService := initServices()

	r := setupRouter(cfg, adapterJWT, userService, userOrderService)

	srv := createServer(cfg, r)

	if err := runServer(ctx, srv); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func initJWTAdapter(supabaseURL string) (*supabase.JWTAdapter, error) {
	adapterJWT, err := supabase.NewJWTAdapter(supabaseURL)
	if err != nil {
		return nil, err
	}
	return adapterJWT, nil
}

func initServices() (*services.UserService, *services.UserOrdersService) {
	userRepo := repository.NewUserRepository(database.DB)
	userService := services.NewUserService(userRepo)

	userOrderRepo := repository.NewUserOrdersRepository(database.DB)
	userOrderService := services.NewUserOrdersService(userOrderRepo)

	return userService, userOrderService
}

func setupRouter(cfg *config.Config, adapterJWT *supabase.JWTAdapter, userService *services.UserService, userOrderService *services.UserOrdersService) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.CORSAllowOrigin},
		AllowMethods: []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))
	r.Use(gin.Logger(), gin.Recovery())

	h := handlers.NewHandler(cfg.SupabaseURL, cfg.SupabaseKey, cfg.SupabaseServiceRoleKey, cfg.ClientURL, adapterJWT, userService, userOrderService)

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
	protected.GET("/users/me/orders", h.GetUserOrders)
	protected.POST("/users/me/orders/:id/complete", h.CompleteUserOrder)
	protected.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from the API"})
	})
	protected.POST("/study/start", handlers.StartStudySessionHandler)
	protected.POST("/study/generate-quiz/:session_id", handlers.CreateQuizFromSession)
	protected.POST("/user/progress", handlers.UpdateProgressHandler(database.DB))

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(handlers.Auth(adapterJWT), h.AdminOnly())
	admin.GET("/users", h.GetAllUsers)
	admin.GET("/users/search", h.GetUserByEmail)
	admin.POST("/users", h.AdminCreateUser)
	admin.DELETE("/users/:id", h.DeleteUser)

	return r
}

func createServer(cfg *config.Config, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func runServer(ctx context.Context, srv *http.Server) error {
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

	return nil
}
