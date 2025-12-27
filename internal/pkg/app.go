// internal/pkg/app.go

package pkg

import (
	"RIP/internal/app/config"
	"RIP/internal/app/handler"
	"fmt"

	"github.com/gin-contrib/cors" // <--- ИМПОРТ 1
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "RIP/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	// --- НАСТРОЙКА CORS (НОВОЕ) ---
	corsConfig := cors.DefaultConfig()
	// Разрешаем все источники (для разработки с Tauri это удобнее всего)
	corsConfig.AllowAllOrigins = true
	// Разрешаем основные методы
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	// Разрешаем заголовки, особенно Authorization (для JWT) и X-Internal-Secret
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Internal-Secret"}
	
	// Применяем middleware
	a.Router.Use(cors.New(corsConfig))
	// ------------------------------

	a.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := a.Router.Group("/api")
	a.Handler.RegisterAPIRoutes(api)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}