package repository

import (
	"RIP/internal/app/ds"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser создает нового пользователя в БД
func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

// GetUserByID находит пользователя по его ID
func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, id).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}

// GetUserByUsername находит пользователя по его логину
func (r *Repository) GetUserByUsername(username string) (ds.User, error) {
	var user ds.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return ds.User{}, errors.New("user not found")
	}
	return user, nil
}

// UpdateUser обновляет данные пользователя (для личного кабинета)
func (r *Repository) UpdateUser(id uint, req ds.UpdateUserRequest) error {
	updates := make(map[string]interface{})

	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password_hash"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		return nil // Нет полей для обновления
	}

	return r.db.Model(&ds.User{}).Where("id = ?", id).Updates(updates).Error
}
