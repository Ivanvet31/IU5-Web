package api

import (
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"lab1/internal/pkg/database"
	"log"

	"github.com/gin-gonic/gin"
)

func StartServer() error {
	log.Println("Server starting...")

	// 1. Подключаемся к базе данных
	db, err := database.ConnectDB()
	if err != nil {
		return err
	}

	// 2. Инициализация зависимостей
	repo := repository.NewRepository(db)
	h := handler.NewHandler(repo)

	// 3. Настройка роутера Gin
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/resources", "./resources")

	// 4. Определяем маршруты (роуты)
	r.GET("/", h.ShowIndexPage)
	r.GET("/strategy/:id", h.ShowStrategyPage)
	// r.GET("/calculator", h.ShowCalculatorPage) // <-- ЭТА СТРОКА УДАЛЕНА

	// --- НОВЫЕ РОУТЫ ДЛЯ ЛАБОРАТОРНОЙ №2 ---
	r.POST("/cart/add/:id", h.AddStrategyToCart)
	r.GET("/cart", h.ShowCartPage)
	r.POST("/cart/delete/:id", h.DeleteRequest)
	r.POST("/cart/update/:id", h.UpdateRequest)

	// 5. Запуск сервера
	log.Println("Server is up and running on port 8080")
	return r.Run()
}
