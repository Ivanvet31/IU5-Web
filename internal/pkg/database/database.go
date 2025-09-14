// internal/pkg/database/database.go
package database

import (
	"fmt"
	"lab1/internal/app" // Импортируем наш пакет с моделями

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB подключается к базе данных и выполняет автомиграцию.
func ConnectDB() (*gorm.DB, error) {
	// Строка подключения (DSN), собранная из данных в docker-compose.yml
	dsn := "host=postgres user=user password=password dbname=recovery_db port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Автоматическая миграция: GORM создаст таблицы на основе ваших Go-структур.
	// Это нужно делать после успешного подключения.
	err = db.AutoMigrate(
		&app.User{},
		&app.Strategy{},
		&app.Request{},
		&app.RequestStrategy{},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка миграции базы данных: %w", err)
	}

	return db, nil
}
