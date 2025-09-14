package api

import (
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"lab1/internal/pkg/database" // Импортируем наш пакет для работы с БД
	"log"

	"github.com/gin-gonic/gin"
)

func StartServer() error {
	log.Println("Server starting...")

	// 1. Подключаемся к базе данных
	db, err := database.ConnectDB()
	if err != nil {
		return err // Если не удалось подключиться, приложение не должно запускаться
	}

	// 2. Инициализация зависимостей с реальным подключением к БД
	repo := repository.NewRepository(db)
	h := handler.NewHandler(repo)

	// 3. Настройка роутера Gin
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/resources", "./resources")

	// 4. Определяем маршруты (роуты)
	r.GET("/", h.ShowIndexPage)
	r.GET("/strategy/:id", h.ShowStrategyPage)
	r.GET("/calculator", h.ShowCalculatorPage)

	// --- НОВЫЕ РОУТЫ ДЛЯ ЛАБОРАТОРНОЙ №2 ---
	r.POST("/cart/add/:id", h.AddStrategyToCart) // Роут для добавления в корзину
	r.GET("/cart", h.ShowCartPage)               // Роут для страницы корзины
	r.POST("/cart/delete/:id", h.DeleteRequest)  // Роут для удаления

	// 5. Запуск сервера
	log.Println("Server is up and running on port 8080")
	return r.Run() // listen and serve on 0.0.0.0:8080
}
