package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/strategies
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
	}

	if err := h.Repository.CreateStrategy(strategy); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, ds.StrategyDTO{
		ID:                strategy.ID,
		Title:             strategy.Title,
		Description:       strategy.Description,
		BaseRecoveryHours: strategy.BaseRecoveryHours,
	})
}

// PUT /api/strategies/:id
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
