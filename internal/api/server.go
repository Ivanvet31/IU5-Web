package api

import (
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
)

func StartServer() error {
	log.Println("Server starting...")

	// Инициализация зависимостей
	repo := repository.NewRepository()
	h := handler.NewHandler(repo)

	// Настройка роутера Gin
	r := gin.Default()

	// Указываем, где лежат HTML-шаблоны
	r.LoadHTMLGlob("templates/*")

	// Указываем, где лежат статические файлы (CSS, изображения)
	r.Static("/resources", "./resources")

	// Определяем маршруты (роуты)
	r.GET("/", h.ShowIndexPage)
	r.GET("/strategy/:id", h.ShowStrategyPage) // Динамический роут для детальной страницы
	r.GET("/calculator", h.ShowCalculatorPage)

	// Запуск сервера
	return r.Run() // listen and serve on 0.0.0.0:8080
}
