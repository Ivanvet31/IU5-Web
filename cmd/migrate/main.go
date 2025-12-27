package main

import (
	"RIP/internal/app/ds"
	"RIP/internal/app/dsn"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Println("Migrations starting...")
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.Strategy{},
		&ds.Recovery_request{},
		&ds.RequestStrategy{},
	)
	if err != nil {
		panic("cant migrate db")
	}
	log.Println("Migrations finished successfully.")
}
