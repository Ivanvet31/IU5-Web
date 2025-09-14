// internal/app/handler/handler.go
package handler

import (
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

func (h *Handler) ShowIndexPage(ctx *gin.Context) {
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
	})
}

func (h *Handler) ShowStrategyPage(ctx *gin.Context) {
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
	})
}

func (h *Handler) ShowCalculatorPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "result_page.html", nil)
}

// --- НОВЫЕ ОБРАБОТЧИКИ ДЛЯ ЛАБОРАТОРНОЙ №2 ---

// AddStrategyToCart добавляет услугу в корзину (черновик заявки)
func (h *Handler) AddStrategyToCart(ctx *gin.Context) {
	// Симулируем, что у нас залогинен пользователь с ID = 1
	const currentUserID uint = 1

	strategyIDStr := ctx.Param("id")
	strategyID, err := strconv.Atoi(strategyIDStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID стратегии")
		return
	}

	// Находим или создаем для пользователя черновик заявки
	request, err := h.Repository.GetOrCreateDraftRequest(currentUserID)
	if err != nil {
		logrus.Errorf("ошибка получения/создания заявки: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	// Добавляем стратегию в заявку. Пока что с фиксированным объемом данных.
	err = h.Repository.AddStrategyToRequest(request.ID, uint(strategyID), 100) // 100GB - для примера
	if err != nil {
		logrus.Errorf("ошибка добавления стратегии в заявку: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	// После успешного добавления перенаправляем пользователя обратно на главную
	ctx.Redirect(http.StatusFound, "/")
}
