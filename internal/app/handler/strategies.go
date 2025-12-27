package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/strategies

// GetStrategies godoc
// @Summary      Получить список стратегий (все)
// @Description  Возвращает список всех активных стратегий с возможностью фильтрации по названию.
// @Tags         strategies
// @Produce      json
// @Param        title query string false "Фильтр по названию стратегии"
// @Success      200 {array} ds.StrategyDTO
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /strategies [get]
func (h *Handler) GetStrategies(c *gin.Context) {
	title := c.Query("title") // Получаем параметр для фильтрации

	strategies, err := h.Repository.GetStrategies(title)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	var strategyDTOs []ds.StrategyDTO
	for _, s := range strategies {
		strategyDTOs = append(strategyDTOs, ds.StrategyDTO{
			ID:                s.ID,
			Title:             s.Title,
			Description:       s.Description,
			ImageURL:          s.ImageURL,
			BaseRecoveryHours: s.BaseRecoveryHours,
		})
	}

	c.JSON(http.StatusOK, strategyDTOs)
}

// GET /api/strategies/:id

// GetStrategy godoc
// @Summary      Получить одну стратегию по ID (все)
// @Description  Возвращает детальную информацию о конкретной стратегии.
// @Tags         strategies
// @Produce      json
// @Param        id path int true "ID стратегии"
// @Success      200 {object} ds.StrategyDTO
// @Failure      404 {object} map[string]string "Стратегия не найдена"
// @Router       /strategies/{id} [get]
func (h *Handler) GetStrategy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	strategy, err := h.Repository.GetStrategyByID(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, ds.StrategyDTO{
		ID:                strategy.ID,
		Title:             strategy.Title,
		Description:       strategy.Description,
		ImageURL:          strategy.ImageURL,
		BaseRecoveryHours: strategy.BaseRecoveryHours,
	})
}

// POST /api/strategies

// CreateStrategy godoc
// @Summary      Создать новую стратегию (модератор)
// @Description  Создает новую запись о стратегии восстановления. Доступно только для модераторов.
// @Tags         strategies
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        strategy body ds.CreateStrategyRequest true "Данные новой стратегии"
// @Success      201 {object} ds.StrategyDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен (не модератор)"
// @Router       /strategies [post]
func (h *Handler) CreateStrategy(c *gin.Context) {
	var req ds.CreateStrategyRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	strategy := &ds.Strategy{
		Title:             req.Title,
		Description:       req.Description,
		BaseRecoveryHours: req.BaseRecoveryHours,
		Status:            "active",
		ImageURL:          req.ImageURL, // Добавлено сохранение ImageURL
	}

	if err := h.Repository.CreateStrategy(strategy); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, ds.StrategyDTO{
		ID:                strategy.ID,
		Title:             strategy.Title,
		Description:       strategy.Description,
		ImageURL:          strategy.ImageURL,
		BaseRecoveryHours: strategy.BaseRecoveryHours,
	})
}

// PUT /api/strategies/:id

// UpdateStrategy godoc
// @Summary      Обновить стратегию (модератор)
// @Description  Обновляет информацию о существующей стратегии.
// @Tags         strategies
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID стратегии"
// @Param        updateData body ds.UpdateStrategyRequest true "Данные для обновления"
// @Success      200 {object} ds.StrategyDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /strategies/{id} [put]
func (h *Handler) UpdateStrategy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.UpdateStrategyRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	updatedStrategy, err := h.Repository.UpdateStrategy(uint(id), req)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, updatedStrategy)
}

// DELETE /api/strategies/:id

// DeleteStrategy godoc
// @Summary      Удалить стратегию (модератор)
// @Description  Удаляет стратегию из системы.
// @Tags         strategies
// @Security     ApiKeyAuth
// @Param        id path int true "ID стратегии для удаления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /strategies/{id} [delete]
func (h *Handler) DeleteStrategy(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.DeleteStrategy(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /api/strategies/:id/image

// UploadStrategyImage godoc
// @Summary      Загрузить изображение для стратегии (модератор)
// @Description  Загружает и привязывает изображение к стратегии.
// @Tags         strategies
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID стратегии"
// @Param        file formData file true "Файл изображения"
// @Success      200 {object} map[string]string "URL загруженного изображения"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /strategies/{id}/image [post]
func (h *Handler) UploadStrategyImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	imageURL, err := h.Repository.UploadStrategyImage(uint(id), file)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}