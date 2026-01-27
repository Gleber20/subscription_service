// @title Subscription Service API
// @version 1.0
// @description Test assignment API for subscriptions
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	_ "subscription_service/docs"
	"subscription_service/internal/config"
	"subscription_service/internal/database"
	httpapi "subscription_service/internal/http"
	"subscription_service/internal/repo/postgres"
	"subscription_service/internal/service"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	_ = godotenv.Load()

	var cfg config.Config
	if err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target:   &cfg,
		Lookuper: envconfig.OsLookuper(),
	}); err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := database.NewPostgres(&cfg)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer db.Close()

	repo := postgres.NewSubscriptionRepo(db)
	svc := service.NewSubscriptionService(repo)
	h := httpapi.NewHandler(svc)

	router := httpapi.NewRouter(h)

	addr := ":" + cfg.HTTPPort

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Запуск сервера в отдельной горутине для graceful shutdown
	go func() {
		log.Printf("subscription_service running on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Ожидаем сигнал завершения через буферизированный канал
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("shutdown signal received...")

	// Даём активным запросам завершиться и мягко закрываем наше приложение
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	log.Println("server stopped gracefully")
}
