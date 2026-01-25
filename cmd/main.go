package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goszakup-automation/internal/api"
	"goszakup-automation/internal/config"
	"goszakup-automation/internal/input"
	"goszakup-automation/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация логгера
	zapLogger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Инициализация Input Service для работы с мышью и клавиатурой
	inputService := input.NewService(zapLogger)

	// Настройка Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API routes
	apiHandler := api.NewHandler(zapLogger, inputService)
	apiGroup := router.Group("/api")
	{
		// Robotogo API endpoints
		testGroup := apiGroup.Group("/robotogo")
		{
			// Мышь
			testGroup.GET("/mouse/position", apiHandler.GetMousePosition)
			testGroup.POST("/mouse/move", apiHandler.MoveMouse)
			testGroup.POST("/mouse/click", apiHandler.Click)
			
			// Клавиатура
			testGroup.POST("/keyboard/type", apiHandler.TypeText)
			
			// Полный цикл (клик + ввод)
			testGroup.POST("/input", apiHandler.InputAtCoordinates)
			
			// Полный цикл: заполнение инпута и клик по кнопке
			testGroup.POST("/fill-and-click", apiHandler.FillInputAndClick)
		}
	}

	// Запуск сервера
	port := cfg.Port
	if port == "" {
		port = "3000"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		zapLogger.Info("Starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}
