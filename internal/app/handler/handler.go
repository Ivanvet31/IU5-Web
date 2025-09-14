package handler

import (
	"lab1/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

// ShowIndexPage теперь также передает информацию о корзине
func (h *Handler) ShowIndexPage(ctx *gin.Context) {
	const currentUserID uint = 1 // Симуляция пользователя

	searchQuery := ctx.Query("query")
	strategies, err := h.Repository.GetStrategies(searchQuery)
	if err != nil {
		logrus.Errorf("ошибка получения стратегий: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	// Получаем черновик заявки, чтобы отобразить виджет корзины
	draftRequest, _ := h.Repository.GetOrCreateDraftRequest(currentUserID)
	// Загружаем связанные стратегии, чтобы посчитать их количество
	cart, _ := h.Repository.GetRequestWithStrategies(draftRequest.ID)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"strategies":  strategies,
		"searchQuery": searchQuery,
		"cart":        cart, // Передаем всю корзину в шаблон
	})
}

// ShowStrategyPage отображает страницу с детальным описанием одной стратегии
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

// ShowCalculatorPage отображает страницу калькулятора
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
		// Не показываем ошибку, если запись уже существует
		// Можно добавить более сложную логику, если нужно
	}

	// После успешного добавления перенаправляем пользователя обратно на главную
	ctx.Redirect(http.StatusFound, "/")
}

// ShowCartPage отображает содержимое корзины
func (h *Handler) ShowCartPage(ctx *gin.Context) {
	const currentUserID uint = 1 // Симуляция пользователя

	draftRequest, err := h.Repository.GetOrCreateDraftRequest(currentUserID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	cart, err := h.Repository.GetRequestWithStrategies(draftRequest.ID)
	if err != nil {
		// Если корзина не найдена (например, после удаления), показываем пустую страницу
		if err == gorm.ErrRecordNotFound {
			ctx.HTML(http.StatusOK, "cart.html", gin.H{"cart": nil})
			return
		}
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	ctx.HTML(http.StatusOK, "cart.html", gin.H{
		"cart": cart,
	})
}

// DeleteRequest выполняет логическое удаление заявки
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

	// После удаления перенаправляем пользователя на пустую корзину
	ctx.Redirect(http.StatusFound, "/cart")
}
