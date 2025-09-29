package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// getCart - вспомогательная функция для получения текущей корзины пользователя
func (h *Handler) getCart(userID uint) ds.Request {
	draftRequest, _ := h.Repository.GetOrCreateDraftRequest(userID)
	if draftRequest.ID == 0 {
		return ds.Request{}
	}
	cart, _ := h.Repository.GetRequestWithStrategies(draftRequest.ID)
	return cart
}

// AddStrategyToCart добавляет услугу в корзину (черновик заявки)
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

	err = h.Repository.AddStrategyToRequest(request.ID, uint(strategyID), 0)
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

	// ИЗМЕНЕНИЕ ЗДЕСЬ: "Распаковываем" указатель для шаблона
	var recoveryTime float64
	if cart.CalculatedRecoveryTimeHours != nil {
		recoveryTime = *cart.CalculatedRecoveryTimeHours // Получаем значение из указателя
	}

	ctx.HTML(http.StatusOK, "cart.html", gin.H{
		"cart":         cart,
		"onCartPage":   true,
		"recoveryTime": recoveryTime, // Передаем в шаблон уже простое число
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

	ctx.Redirect(http.StatusFound, "/cart")
}

// UpdateRequest сохраняет данные из формы калькулятора и запускает расчет
func (h *Handler) UpdateRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}
	requestID := uint(id)

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

	updateData := ds.Request{
		ItSkillLevel:         &skillLevel,
		DocumentationQuality: &docQuality,
		NetworkBandwidthMbps: bandwidth,
	}

	if err := h.Repository.UpdateRequestDetails(requestID, updateData); err != nil {
		logrus.Errorf("ошибка обновления заявки: %v", err)
		ctx.String(http.StatusInternalServerError, "Ошибка сервера")
		return
	}

	cart, _ := h.Repository.GetRequestWithStrategies(requestID)
	var associations []ds.RequestStrategy
	for _, strategy := range cart.Strategies {
		formFieldName := "data_gb_" + strconv.Itoa(int(strategy.ID))
		dataGbStr := ctx.PostForm(formFieldName)

		dataGb, err := strconv.Atoi(dataGbStr)
		if err != nil {
			dataGb = 0
		}

		if err := h.Repository.UpdateRequestStrategyData(requestID, strategy.ID, dataGb); err != nil {
			logrus.Errorf("ошибка обновления данных для стратегии %d: %v", strategy.ID, err)
		}

		associations = append(associations, ds.RequestStrategy{DataToRecoverGB: dataGb})
	}

	updatedRequest, _ := h.Repository.GetRequestWithStrategies(requestID)
	if err := h.Repository.CalculateAndSaveRecoveryTime(updatedRequest, associations); err != nil {
		logrus.Errorf("ошибка расчета времени: %v", err)
	}

	ctx.Redirect(http.StatusFound, "/cart")
}

func (h *Handler) ShowRequestByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}

	// Используем существующий метод, который умеет получать заявку со стратегиями по ID
	request, err := h.Repository.GetRequestWithStrategies(uint(id))
	if err != nil {
		// Здесь можно обработать ошибку "не найдено" отдельно, если нужно
		logrus.Errorf("заявка с ID %d не найдена: %v", id, err)
		ctx.String(http.StatusNotFound, "Заявка не найдена")
		return
	}

	ctx.HTML(http.StatusOK, "cart.html", gin.H{
		"cart": request, // Передаем найденную заявку в ту же переменную "cart"
	})
}
