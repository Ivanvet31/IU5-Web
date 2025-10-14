package handler

import (
	"RIP/internal/app/config"
	"RIP/internal/app/redis"
	"RIP/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWTConfig  *config.JWTConfig
}

func NewHandler(r *repository.Repository, redis *redis.Client, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redis,
		JWTConfig:  jwtConfig,
	}
}

func (h *Handler) RegisterAPIRoutes(api *gin.RouterGroup) {
	// --- Публичные эндпоинты (доступны всем) ---
	api.POST("/users", h.Register)
	api.POST("/auth/login", h.Login)
	api.GET("/strategies", h.GetStrategies)
	api.GET("/strategies/:id", h.GetStrategy)

	// --- Группа эндпоинтов, требующих авторизации ---
	auth := api.Group("/")
	auth.Use(h.AuthMiddleware)
	{
		// Пользователи
		auth.POST("/auth/logout", h.Logout)
		auth.GET("/users/me", h.GetMyUserData)
		auth.PUT("/users/me", h.UpdateMyUserData)

		// Заявки
		requests := auth.Group("/recovery_requests")
		{
			requests.GET("/cart", h.GetCartBadge)
			requests.GET("", h.ListRequests)
			requests.POST("/draft/strategies/:strategy_id", h.AddStrategyToDraft)

			requestByID := requests.Group("/:id")
			{
				requestByID.GET("", h.GetRequest)
				requestByID.PUT("", h.UpdateRequestDetails)
				requestByID.DELETE("", h.DeleteRequest)
				requestByID.PUT("/form", h.FormRequest)

				// М-М связи
				strategiesInRequest := requestByID.Group("/strategies/:strategy_id")
				{
					strategiesInRequest.PUT("", h.UpdateRequestStrategy)
					strategiesInRequest.DELETE("", h.RemoveStrategyFromRequest)
				}
			}
		}
	}

	// --- Группа эндпоинтов, требующих прав модератора ---
	moderator := api.Group("/")
	moderator.Use(h.AuthMiddleware, h.ModeratorMiddleware)
	{
		// Управление услугами
		moderator.POST("/strategies", h.CreateStrategy)
		moderator.PUT("/strategies/:id", h.UpdateStrategy)
		moderator.DELETE("/strategies/:id", h.DeleteStrategy)
		moderator.POST("/strategies/:id/image", h.UploadStrategyImage)

		// Управление заявками
		moderator.PUT("/recovery_requests/:id/resolve", h.ResolveRequest)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{"error": err.Error()})
}
