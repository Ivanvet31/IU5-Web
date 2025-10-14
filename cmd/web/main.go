package main

import (
	"RIP/internal/app/config"
	"RIP/internal/app/dsn"
	"RIP/internal/app/handler"
	"RIP/internal/app/redis"
	"RIP/internal/app/repository"
	"RIP/internal/pkg"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// @title           API Системы восстановления
// @version         1.0
// @description     API-сервер для управления заявками и стратегиями восстановления данных.
// @contact.name    API Support
// @contact.email   support@recovery-api.com
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	logrus.Info("Application starting...")

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading .env file: %v", err)
	}

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	if postgresString == "" {
		logrus.Fatal("DSN string is empty, check your .env file")
	}

	repo, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	redisClient, err := redis.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}

	hand := handler.NewHandler(repo, redisClient, &conf.JWT)
	router := gin.Default()
	application := pkg.NewApp(conf, router, hand)

	application.RunApp()
}
