package handler

import (
	"lab1/internal/app/models"
	"lab1/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// getCart - вспомогательная функция
func (h *Handler) getCart(userID uint) models.Request {
	draftRequest, _ := h.Repository.GetOrCreateDraftRequest(userID)
	if draftRequest.ID == 0 {
		return models.Request{}
	}
	cart, _ := h.Repository.GetRequestWithStrategies(draftRequest.ID)
	return cart
}

func (h *Handler) ShowIndexPage(ctx *gin.Context) {
	const currentUserID uint = 1
	searchQuery := ctx.Query("query")

	strategies, err := h.Repository.GetStrategies(searchQuery)
	if err != nil {
		logrus.Errorf("ошибка получения стратегий: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"strategies":  strategies,
		"searchQuery": searchQuery,
		"cart":        h.getCart(currentUserID),
	})
}

func (h *Handler) ShowStrategyPage(ctx *gin.Context) {
	const currentUserID uint = 1
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Errorf("некорректный ID: %v", err)
		ctx.String(http.StatusBadRequest, "Некорректный ID")
		return
	}

	strategy, err := h.Repository.GetStrategyByID(uint(id))
	if err != nil {
		logrus.Errorf("стратегия не найдена: %v", err)
		ctx.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	ctx.HTML(http.StatusOK, "description.html", gin.H{
		"strategy": strategy,
		"cart":     h.getCart(currentUserID),
	})
}

// ФУНКЦИЯ ShowCalculatorPage БЫЛА ПОЛНОСТЬЮ УДАЛЕНА

func (h *Handler) AddStrategyToCart(ctx *gin.Context) {
	const currentUserID uint = 1
	strategyIDStr := ctx.Param("id")
	strategyID, err := strconv.Atoi(strategyIDStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID стратегии")
		return
	}

	request, err := h.Repository.GetOrCreateDraftRequest(currentUserID)
	if err != nil {
		logrus.Errorf("ошибка получения/создания заявки: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	err = h.Repository.AddStrategyToRequest(request.ID, uint(strategyID), 100)
	if err != nil {
		logrus.Warnf("ошибка добавления стратегии в заявку (возможно, уже существует): %v", err)
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) ShowCartPage(ctx *gin.Context) {
	const currentUserID uint = 1
	cart := h.getCart(currentUserID)

	if cart.ID == 0 {
		ctx.HTML(http.StatusOK, "cart.html", gin.H{
			"cart":       nil,
			"onCartPage": true,
		})
		return
	}

	ctx.HTML(http.StatusOK, "cart.html", gin.H{
		"cart":       cart,
		"onCartPage": true,
	})
}

func (h *Handler) DeleteRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}

	err = h.Repository.LogicallyDeleteRequest(uint(id))
	if err != nil {
		logrus.Errorf("ошибка удаления заявки: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	ctx.Redirect(http.StatusFound, "/cart")
}

func (h *Handler) UpdateRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}

	skillLevel := ctx.PostForm("it_skill_level")
	docQuality := ctx.PostForm("documentation_quality")
	bandwidthStr := ctx.PostForm("network_bandwidth_mbps")

	var bandwidth *int
	if bandwidthStr != "" {
		val, err := strconv.Atoi(bandwidthStr)
		if err == nil {
			bandwidth = &val
		} else {
			logrus.Warnf("Некорректное значение пропускной способности: %v", bandwidthStr)
		}
	}

	updateData := models.Request{
		ItSkillLevel:         &skillLevel,
		DocumentationQuality: &docQuality,
		NetworkBandwidthMbps: bandwidth,
	}

	if err := h.Repository.UpdateRequestDetails(uint(id), updateData); err != nil {
		logrus.Errorf("ошибка обновления заявки: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	ctx.Redirect(http.StatusFound, "/cart")
}
