package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/requests/cart
func (h *Handler) GetCartBadge(c *gin.Context) {
	draft, err := h.Repository.GetOrCreateDraftRequest(hardcodedUserID)
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

// GET /api/requests
func (h *Handler) ListRequests(c *gin.Context) {
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	requests, err := h.Repository.ListRequestsFiltered(status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, requests)
}

// GET /api/requests/:id
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

	// Маппинг в DTO для чистого ответа
	var strategyDTOs []ds.StrategyDTO
	for _, s := range request.Strategies {
		strategyDTOs = append(strategyDTOs, ds.StrategyDTO{
			ID:                s.ID,
			Title:             s.Title,
			Description:       s.Description,
			ImageURL:          s.ImageURL,
			BaseRecoveryHours: s.BaseRecoveryHours,
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

// PUT /api/requests/:id
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

// DELETE /api/requests/:id
func (h *Handler) DeleteRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.LogicallyDeleteRequest(uint(id), hardcodedUserID); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /api/requests/draft/strategies/:strategy_id
func (h *Handler) AddStrategyToDraft(c *gin.Context) {
	strategyID, err := strconv.Atoi(c.Param("strategy_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	draft, err := h.Repository.GetOrCreateDraftRequest(hardcodedUserID)
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

// PUT /api/requests/:id/form
func (h *Handler) FormRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.FormRequest(uint(id), hardcodedUserID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err) // 400, т.к. ошибка скорее всего в бизнес-логике (не тот статус и т.д.)
		return
	}
	c.Status(http.StatusNoContent)
}

// PUT /api/requests/:id/resolve
func (h *Handler) ResolveRequest(c *gin.Context) {
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
	// В реальном приложении ID модератора берется из токена
	if err := h.Repository.ResolveRequest(uint(id), hardcodedUserID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PUT /api/requests/:request_id/strategies/:strategy_id
func (h *Handler) UpdateRequestStrategy(c *gin.Context) {
	requestID, err := strconv.Atoi(c.Param("request_id"))
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

// DELETE /api/requests/:request_id/strategies/:strategy_id
func (h *Handler) RemoveStrategyFromRequest(c *gin.Context) {
	requestID, err := strconv.Atoi(c.Param("request_id"))
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
