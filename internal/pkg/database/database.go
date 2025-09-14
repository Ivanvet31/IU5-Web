// internal/pkg/database/database.go
package database

import (
	"fmt"
	"lab1/internal/app/models" // Импортируем наш пакет с моделями

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB подключается к базе данных и выполняет автомиграцию.
func ConnectDB() (*gorm.DB, error) {
	// Строка подключения (DSN), собранная из данных в docker-compose.yml
	// host=postgres - важно, что мы обращаемся к контейнеру по имени сервиса
	dsn := "host=localhost user=user password=password dbname=recovery_db port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}

	// Автоматическая миграция: GORM создаст/обновит таблицы на основе Go-структур.
	err = db.AutoMigrate(
		&models.User{},
		&models.Strategy{},
		&models.Request{},
		&models.RequestStrategy{},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка миграции базы данных: %w", err)
	}

	fmt.Println("Подключение к базе данных и миграция прошли успешно!")
	return db, nil
}
