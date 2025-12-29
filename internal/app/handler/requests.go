package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCartBadge godoc
// @Summary      Получить информацию для иконки корзины (авторизованные пользователи)
// @Description  Возвращает ID черновика текущего пользователя и количество стратегий в нем.
// @Tags         requests
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} ds.CartBadgeDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /recovery_requests/cart [get]
func (h *Handler) GetCartBadge(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{RequestID: nil, Count: 0})
		return
	}

	draft, err := h.Repository.GetOrCreateDraftRequest(userID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{RequestID: nil, Count: 0})
		return
	}

	fullRequest, err := h.Repository.GetRequestWithStrategies(draft.ID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{RequestID: &draft.ID, Count: 0})
		return
	}

	c.JSON(http.StatusOK, ds.CartBadgeDTO{
		RequestID: &fullRequest.ID,
		Count:     len(fullRequest.Strategies),
	})
}

// ListRequests godoc
// @Summary      Получить список заявок (авторизованные пользователи)
// @Description  Возвращает список заявок. Для модератора - все, для пользователя - только его.
// @Tags         requests
// @Produce      json
// @Security     ApiKeyAuth
// @Param        status query string false "Фильтр по статусу заявки (formed, completed, rejected)"
// @Param        from query string false "Фильтр по дате 'от' (формат YYYY-MM-DD)"
// @Param        to query string false "Фильтр по дате 'до' (формат YYYY-MM-DD)"
// @Success      200 {array} ds.RequestDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /recovery_requests [get]
func (h *Handler) ListRequests(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}
	isModerator := isUserModerator(c)

	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	requests, err := h.Repository.ListRequestsFiltered(userID, isModerator, status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, requests)
}

// GetRequest godoc
// @Summary      Получить одну заявку по ID (авторизованные пользователи)
// @Description  Возвращает полную информацию о заявке, включая привязанные стратегии.
// @Tags         requests
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      200 {object} ds.RequestDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      404 {object} map[string]string "Заявка не найдена"
// @Router       /recovery_requests/{id} [get]
func (h *Handler) GetRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.GetRequestWithStrategies(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	var strategyDTOs []ds.StrategyDTO
	for _, s := range request.Strategies {
		// --- НОВАЯ ЛОГИКА: Получаем объем данных из связующей таблицы ---
		var assoc ds.RequestStrategy
		// Ищем запись в request_strategies по ID заявки и ID стратегии
		h.Repository.DB().Where("request_id = ? AND strategy_id = ?", request.ID, s.ID).First(&assoc)
		
		strategyDTOs = append(strategyDTOs, ds.StrategyDTO{
			ID:                s.ID,
			Title:             s.Title,
			Description:       s.Description,
			ImageURL:          s.ImageURL,
			BaseRecoveryHours: s.BaseRecoveryHours,
			DataToRecoverGB:   assoc.DataToRecoverGB, // Заполняем поле
		})
	}

	response := ds.RequestDTO{
		ID:                          request.ID,
		Status:                      request.Status,
		CreatedAt:                   request.CreatedAt,
		CreatorUsername:             request.User.Username,
		ItSkillLevel:                request.ItSkillLevel,
		NetworkBandwidthMbps:        request.NetworkBandwidthMbps,
		DocumentationQuality:        request.DocumentationQuality,
		CalculatedRecoveryTimeHours: request.CalculatedRecoveryTimeHours,
		Strategies:                  strategyDTOs,
	}

	if request.Moderator != nil {
		response.ModeratorUsername = &request.Moderator.Username
	}

	c.JSON(http.StatusOK, response)
}

// UpdateRequestDetails godoc
// @Summary      Обновить детали заявки (авторизованные пользователи)
// @Description  Позволяет пользователю обновить поля своей заявки (уровень навыков, пропускная способность сети и т.д.).
// @Tags         requests
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        updateData body ds.UpdateRequestDetailsRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /recovery_requests/{id} [put]
func (h *Handler) UpdateRequestDetails(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	var req ds.UpdateRequestDetailsRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateRequestDetails(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteRequest godoc
// @Summary      Удалить заявку (черновик) (авторизованные пользователи)
// @Description  Логически удаляет заявку, переводя ее в статус "удалена". Доступно только для создателя и только для черновиков.
// @Tags         requests
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /recovery_requests/{id} [delete]
func (h *Handler) DeleteRequest(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.LogicallyDeleteRequest(uint(id), userID); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AddStrategyToDraft godoc
// @Summary      Добавить стратегию в черновик заявки (авторизованные пользователи)
// @Description  Находит или создает черновик заявки для текущего пользователя и добавляет в него стратегию.
// @Tags         strategies
// @Security     ApiKeyAuth
// @Param        strategy_id path int true "ID стратегии для добавления"
// @Success      201 "Created"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /recovery_requests/draft/strategies/{strategy_id} [post]
func (h *Handler) AddStrategyToDraft(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	strategyID, err := strconv.Atoi(c.Param("strategy_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	draft, err := h.Repository.GetOrCreateDraftRequest(userID)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	if err := h.Repository.AddStrategyToRequest(draft.ID, uint(strategyID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusCreated)
}

// FormRequest godoc
// @Summary      Сформировать заявку (авторизованные пользователи)
// @Description  Переводит заявку из статуса "черновик" в "сформирована".
// @Tags         requests
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки (черновика)"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /recovery_requests/{id}/form [put]
func (h *Handler) FormRequest(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.FormRequest(uint(id), userID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ResolveRequest godoc
// @Summary      Завершить или отклонить заявку (модератор)
// @Description  Модератор завершает (с расчетом) или отклоняет заявку. Расчет выполняется асинхронно в Python-сервисе.
// @Tags         requests
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        action body ds.ResolveRequest true "Действие: 'complete' или 'reject'"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /recovery_requests/{id}/resolve [put]
func (h *Handler) ResolveRequest(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	var req ds.ResolveRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.ResolveRequest(uint(id), userID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// UpdateRequestStrategy godoc
// @Summary      Обновить данные стратегии в заявке (авторизованные пользователи)
// @Description  Изменяет дополнительные данные (например, объем данных) для конкретной стратегии в рамках одной заявки.
// @Tags         m-m
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        strategy_id path int true "ID стратегии"
// @Param        updateData body ds.UpdateRequestStrategyRequest true "Новые данные"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /recovery_requests/{id}/strategies/{strategy_id} [put]
func (h *Handler) UpdateRequestStrategy(c *gin.Context) {
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	strategyID, err := strconv.Atoi(c.Param("strategy_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	var req ds.UpdateRequestStrategyRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.UpdateRequestStrategyData(uint(requestID), uint(strategyID), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveStrategyFromRequest godoc
// @Summary      Удалить стратегию из заявки (авторизованные пользователи)
// @Description  Удаляет связь между заявкой и стратегией.
// @Tags         m-m
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        strategy_id path int true "ID стратегии"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /recovery_requests/{id}/strategies/{strategy_id} [delete]
func (h *Handler) RemoveStrategyFromRequest(c *gin.Context) {
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	strategyID, err := strconv.Atoi(c.Param("strategy_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.RemoveStrategyFromRequest(uint(requestID), uint(strategyID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// SetRequestResult godoc
// @Summary      Установить результат расчета (внутренний метод)
// @Description  Принимает результат от асинхронного сервиса. Защищен секретным заголовком.
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        id path int true "ID заявки"
// @Param        result body ds.AsyncCalculateResult true "Результат"
// @Success      200 {string} string "OK"
// @Router       /internal/requests/{id}/result [put]
func (h *Handler) SetRequestResult(c *gin.Context) {
	// 1. Псевдо-авторизация по ключу
	secretKey := c.GetHeader("X-Internal-Secret")
	if secretKey != "my_secret_8_byte_key" { // Тот же ключ, что в Python
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	// 2. Получение ID
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	// 3. Парсинг тела
	var req ds.AsyncCalculateResult
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	// 4. Обновление в БД
	if err := h.Repository.UpdateRecoveryTime(uint(id), req.CalculatedTime); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}