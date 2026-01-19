// @title Subscription Service API
// @version 1.0
// @description Test assignment API for subscriptions
// @BasePath /
package main

import (
	"context"
	"fmt"
	"log"
	_ "subscription_service/docs"
	"subscription_service/internal/config"
	"subscription_service/internal/database"
	"subscription_service/internal/http"
	"subscription_service/internal/repo/postgres"
	"subscription_service/internal/service"

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
	h := http.NewHandler(svc)

	r := http.NewRouter(h)

	addr := ":" + cfg.HTTPPort
	fmt.Println("subscription_service running on", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
