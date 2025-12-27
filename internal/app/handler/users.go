package handler

import (
	"RIP/internal/app/ds"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// POST /api/users

// Register godoc
// @Summary      Регистрация нового пользователя (гость)
// @Description  Создает нового пользователя в системе. По умолчанию роль "пользователь", не "модератор".
// @Tags         authorization
// @Accept       json
// @Produce      json
// @Param        credentials body ds.UserRegisterRequest true "Данные для регистрации"
// @Success      201 {object} ds.UserDTO "Пользователь успешно создан"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /users [post]
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

// GetMyUserData godoc
// @Summary      Получить данные текущего пользователя (авторизованные пользователи)
// @Description  Возвращает информацию о пользователе, чей токен используется для запроса.
// @Tags         users
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} ds.UserDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Router       /users/me [get]
func (h *Handler) GetMyUserData(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}
	// В реальном приложении ID берется из контекста после middleware
	user, err := h.Repository.GetUserByID(userID)
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

// UpdateMyUserData godoc
// @Summary      Обновить данные текущего пользователя (авторизованные пользователи)
// @Description  Позволяет авторизованному пользователю обновить свой логин или пароль.
// @Tags         users
// @Accept       json
// @Security     ApiKeyAuth
// @Param        updateData body ds.UpdateUserRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /users/me [put]
func (h *Handler) UpdateMyUserData(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	var req ds.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateUser(userID, req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /api/auth/login

// Login godoc
// @Summary      Аутентификация пользователя (гость)
// @Description  Получение JWT токена по логину и паролю для доступа к защищенным эндпоинтам.
// @Tags         authorization
// @Accept       json
// @Produce      json
// @Param        credentials body ds.UserLoginRequest true "Учетные данные"
// @Success      200 {object} ds.LoginResponse
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Неверные учетные данные"
// @Router       /auth/login [post]
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

	claims := ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.JWTConfig.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:      user.ID,
		Username:    user.Username,
		IsModerator: user.IsModerator,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.JWTConfig.Secret))
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, ds.LoginResponse{
		Token: tokenString, // --- ИСПОЛЬЗОВАТЬ СГЕНЕРИРОВАННЫЙ ТОКЕН ---
		User: ds.UserDTO{
			ID:          user.ID,
			Username:    user.Username,
			IsModerator: user.IsModerator,
		},
	})
}

// POST /api/auth/logout

// Logout godoc
// @Summary      Выход из системы (авторизованные пользователи)
// @Description  Добавляет текущий JWT токен в черный список, делая его недействительным.
// @Tags         authorization
// @Security     ApiKeyAuth
// @Success      200 {object} map[string]string "Сообщение об успехе"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		h.errorHandler(c, http.StatusBadRequest, errors.New("invalid header"))
		return
	}
	tokenStr := authHeader[len("Bearer "):]

	err := h.Redis.WriteJWTToBlacklist(c.Request.Context(), tokenStr, h.JWTConfig.ExpiresIn)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Деавторизация прошла успешно",
	})
}
