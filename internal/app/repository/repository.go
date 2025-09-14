// internal/app/repository/repository.go
package repository

import (
	"lab1/internal/app" // Импортируем наши модели

	"gorm.io/gorm"
)

// Repository теперь содержит подключение к БД
type Repository struct {
	db *gorm.DB
}

// NewRepository теперь принимает подключение к БД
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetStrategies теперь выполняет запрос к БД через GORM
func (r *Repository) GetStrategies(query string) ([]app.Strategy, error) {
	var strategies []app.Strategy
	// Строим запрос. ILIKE - это регистронезависимый поиск в PostgreSQL
	tx := r.db.Where("status = ?", "active")
	if query != "" {
		tx = tx.Where("title ILIKE ?", "%"+query+"%")
	}
	err := tx.Find(&strategies).Error
	return strategies, err
}

// GetStrategyByID ищет одну запись в БД
func (r *Repository) GetStrategyByID(id uint) (app.Strategy, error) {
	var strategy app.Strategy
	err := r.db.First(&strategy, id).Error
	return strategy, err
}

// --- НОВЫЕ МЕТОДЫ ДЛЯ РАБОТЫ С ЗАЯВКАМИ (КОРЗИНОЙ) ---

// GetOrCreateDraftRequest находит или создает черновик заявки для пользователя
func (r *Repository) GetOrCreateDraftRequest(userID uint) (app.Request, error) {
	var request app.Request
	// Ищем заявку со статусом 'draft' для данного пользователя.
	// Если не найдена - создаем новую.
	err := r.db.Where(app.Request{UserID: userID, Status: "draft"}).FirstOrCreate(&request).Error
	return request, err
}

// AddStrategyToRequest добавляет стратегию в заявку
func (r *Repository) AddStrategyToRequest(requestID, strategyID uint, dataGB int) error {
	association := app.RequestStrategy{
		RequestID:       requestID,
		StrategyID:      strategyID,
		DataToRecoverGB: dataGB, // Используем данные для расчета
	}
	// GORM обработает конфликт составного первичного ключа и не даст добавить дубликат
	return r.db.Create(&association).Error
}

// LogicallyDeleteRequest выполняет ЧИСТЫЙ SQL-ЗАПРОС, как требует задание
func (r *Repository) LogicallyDeleteRequest(requestID uint) error {
	// Используем db.Exec для выполнения сырого SQL
	result := r.db.Exec("UPDATE requests SET status = ? WHERE id = ?", "deleted", requestID)
	return result.Error
}
