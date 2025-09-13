package api

import (
	"lab1/internal/app/handler"
	"lab1/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
)

func StartServer() error {
	log.Println("Server starting...")

	// 1. Инициализация зависимостей
	repo := repository.NewRepository()
	h := handler.NewHandler(repo)

	// 2. Настройка роутера Gin
	r := gin.Default()

	// 3. Указываем, где лежат HTML-шаблоны
	r.LoadHTMLGlob("templates/*")

	// 4. Указываем, где лежат статические файлы (CSS, изображения)
	// Теперь URL /resources/styles/style.css будет искать файл в папке ./resources/styles/style.css
	r.Static("/resources", "./resources")

	// 5. Определяем маршруты (роуты)
	r.GET("/", h.ShowIndexPage)
	r.GET("/strategy/:id", h.ShowStrategyPage) // Динамический роут для детальной страницы
	r.GET("/calculator", h.ShowCalculatorPage)

	// 6. Запуск сервера
	return r.Run() // listen and serve on 0.0.0.0:8080
}
