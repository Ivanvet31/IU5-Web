package repository

import (
	"lab1/internal/app/models" // Используем новый пакет models

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
func (r *Repository) GetStrategies(query string) ([]models.Strategy, error) {
	var strategies []models.Strategy
	tx := r.db.Where("status = ?", "active")
	if query != "" {
		tx = tx.Where("title ILIKE ?", "%"+query+"%")
	}
	err := tx.Find(&strategies).Error
	return strategies, err
}

// GetStrategyByID ищет одну запись в БД
func (r *Repository) GetStrategyByID(id uint) (models.Strategy, error) {
	var strategy models.Strategy
	err := r.db.First(&strategy, id).Error
	return strategy, err
}

// --- НОВЫЕ МЕТОДЫ ДЛЯ РАБОТЫ С ЗАЯВКАМИ ---

func (r *Repository) GetOrCreateDraftRequest(userID uint) (models.Request, error) {
	var request models.Request
	err := r.db.Where(models.Request{UserID: userID, Status: "draft"}).FirstOrCreate(&request).Error
	return request, err
}

func (r *Repository) AddStrategyToRequest(requestID, strategyID uint, dataGB int) error {
	association := models.RequestStrategy{
		RequestID:       requestID,
		StrategyID:      strategyID,
		DataToRecoverGB: dataGB,
	}
	return r.db.Create(&association).Error
}

func (r *Repository) GetRequestWithStrategies(requestID uint) (models.Request, error) {
	var request models.Request
	err := r.db.Preload("Strategies").First(&request, requestID).Error
	return request, err
}

func (r *Repository) LogicallyDeleteRequest(requestID uint) error {
	result := r.db.Exec("UPDATE requests SET status = ? WHERE id = ?", "deleted", requestID)
	return result.Error
}
