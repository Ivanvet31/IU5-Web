package repository

import (
	"RIP/internal/app/ds"
)

func (r *Repository) GetStrategies(query string) ([]ds.Strategy, error) {
	var strategies []ds.Strategy
	tx := r.db.Where("status = ?", "active")
	if query != "" {
		tx = tx.Where("title ILIKE ?", "%"+query+"%")
	}
	err := tx.Find(&strategies).Error
	return strategies, err
}

func (r *Repository) GetStrategyByID(id uint) (ds.Strategy, error) {
	var strategy ds.Strategy
	err := r.db.First(&strategy, id).Error
	return strategy, err
}
