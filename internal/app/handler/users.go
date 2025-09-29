package handler

import (
	"RIP/internal/app/ds"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// POST /api/users
func (h *Handler) Register(c *gin.Context) {
	var req ds.UserRegisterRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	user := ds.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		IsModerator:  false, // По умолчанию пользователь не модератор
		IsActive:     true,
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, ds.UserDTO{
		ID:          user.ID,
		Username:    user.Username,
		IsModerator: user.IsModerator,
	})
}

// GET /api/users/me (Предполагается, что ID пользователя будет в токене)
func (h *Handler) GetMyUserData(c *gin.Context) {
	// В реальном приложении ID берется из контекста после middleware
	user, err := h.Repository.GetUserByID(hardcodedUserID)
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, ds.UserDTO{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		IsModerator: user.IsModerator,
	})
}

// PUT /api/users/me
func (h *Handler) UpdateMyUserData(c *gin.Context) {
	var req ds.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateUser(hardcodedUserID, req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /api/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req ds.UserLoginRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByUsername(req.Username)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	c.JSON(http.StatusOK, ds.LoginResponse{
		Token: "super_secret_jwt_token", // В реальном приложении здесь генерируется JWT
		User: ds.UserDTO{
			ID:          user.ID,
			Username:    user.Username,
			IsModerator: user.IsModerator,
		},
	})
}

// POST /api/auth/logout
func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}
