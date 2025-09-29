package handler

import (
	"RIP/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/", h.ShowIndexPage)
	router.GET("/strategy/:id", h.ShowStrategyPage)
	router.GET("/cart", h.ShowCartPage)

	router.POST("/cart/add/:id", h.AddStrategyToCart)
	router.POST("/cart/delete/:id", h.DeleteRequest)
	router.POST("/cart/update/:id", h.UpdateRequest)

	router.GET("/:id", h.ShowRequestByID)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/resources", "./resources")
}
