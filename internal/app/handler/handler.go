package handler

import (
	"RIP/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const hardcodedUserID = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterAPIRoutes(api *gin.RouterGroup) {
	// --- Домен услуг (стратегий) ---
	strategies := api.Group("/strategies")
	{
		strategies.GET("", h.GetStrategies)
		strategies.GET("/:id", h.GetStrategy)
		strategies.POST("", h.CreateStrategy)
		strategies.PUT("/:id", h.UpdateStrategy)
		strategies.DELETE("/:id", h.DeleteStrategy)
		strategies.POST("/:id/image", h.UploadStrategyImage)
	}

	// --- Домен заявок (requests) ---
	requests := api.Group("/requests")
	{
		// Общие роуты для заявок
		requests.GET("/cart", h.GetCartBadge)
		requests.GET("", h.ListRequests)
		requests.POST("/draft/strategies/:strategy_id", h.AddStrategyToDraft)

		// Группа для действий с конкретной заявкой по ее ID
		requestByID := requests.Group("/:id")
		{
			requestByID.GET("", h.GetRequest)
			requestByID.PUT("", h.UpdateRequestDetails)
			requestByID.DELETE("", h.DeleteRequest)
			requestByID.PUT("/form", h.FormRequest)
			requestByID.PUT("/resolve", h.ResolveRequest)

			// Группа для действий со стратегиями ВНУТРИ конкретной заявки
			strategiesInRequest := requestByID.Group("/strategies/:strategy_id")
			{
				strategiesInRequest.PUT("", h.UpdateRequestStrategy)
				strategiesInRequest.DELETE("", h.RemoveStrategyFromRequest)
			}
		}
	}

	// --- Домен пользователей и аутентификации ---
	users := api.Group("/users")
	{
		users.POST("", h.Register)
		users.GET("/me", h.GetMyUserData)
		users.PUT("/me", h.UpdateMyUserData)
	}
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
	}
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{"error": err.Error()})
}
