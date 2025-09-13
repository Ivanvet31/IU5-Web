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

// ShowIndexPage ТЕПЕРЬ ОБРАБАТЫВАЕТ ПОИСК
func (h *Handler) ShowIndexPage(ctx *gin.Context) {
	// Получаем значение параметра "query" из URL. Если его нет, будет пустая строка.
	searchQuery := ctx.Query("query")

	// Передаем поисковый запрос в репозиторий
	strategies, err := h.Repository.GetStrategies(searchQuery)
	if err != nil {
		logrus.Errorf("ошибка получения стратегий: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	// Передаем в шаблон не только стратегии, но и сам поисковый запрос
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"strategies":  strategies,
		"searchQuery": searchQuery, // Это нужно для сохранения текста в поле поиска
	})
}

// ShowStrategyPage остается без изменений
func (h *Handler) ShowStrategyPage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Errorf("некорректный ID: %v", err)
		ctx.String(http.StatusBadRequest, "Некорректный ID")
		return
	}

	strategy, err := h.Repository.GetStrategyByID(id)
	if err != nil {
		logrus.Errorf("стратегия не найдена: %v", err)
		ctx.String(http.StatusNotFound, "Страница не найдена")
		return
	}

	ctx.HTML(http.StatusOK, "description.html", gin.H{
		"strategy": strategy,
	})
}

// ShowCalculatorPage остается без изменений
func (h *Handler) ShowCalculatorPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "result_page.html", nil)
}
