package repository

import (
	"RIP/internal/app/ds"
	"strings"

	"gorm.io/gorm"
)

func (r *Repository) GetOrCreateDraftRequest(userID uint) (ds.Request, error) {
	var request ds.Request
	err := r.db.Where(ds.Request{UserID: userID, Status: "draft"}).FirstOrCreate(&request).Error
	return request, err
}

func (r *Repository) AddStrategyToRequest(requestID, strategyID uint, dataGB int) error {
	var existing ds.RequestStrategy
	err := r.db.Where("request_id = ? AND strategy_id = ?", requestID, strategyID).First(&existing).Error

	if err == nil {
		return nil // Already exists, no error
	}

	if err != gorm.ErrRecordNotFound {
		return err // Another DB error
	}

	association := ds.RequestStrategy{
		RequestID:       requestID,
		StrategyID:      strategyID,
		DataToRecoverGB: dataGB,
	}
	return r.db.Create(&association).Error
}

func (r *Repository) GetRequestWithStrategies(requestID uint) (ds.Request, error) {
	var request ds.Request
	err := r.db.Preload("Strategies").First(&request, requestID).Error
	return request, err
}

func (r *Repository) LogicallyDeleteRequest(requestID uint) error {
	return r.db.Exec("UPDATE requests SET status = ? WHERE id = ?", "deleted", requestID).Error
}

func (r *Repository) UpdateRequestDetails(requestID uint, details ds.Request) error {
	return r.db.Model(&ds.Request{}).Where("id = ?", requestID).Updates(details).Error
}

func (r *Repository) UpdateRequestStrategyData(requestID uint, strategyID uint, dataGB int) error {
	return r.db.Model(&ds.RequestStrategy{}).
		Where("request_id = ? AND strategy_id = ?", requestID, strategyID).
		Update("data_to_recover_gb", dataGB).Error
}

func (r *Repository) CalculateAndSaveRecoveryTime(request ds.Request, associations []ds.RequestStrategy) error {
	var totalBaseHours float64 = 0
	var totalDataGB int = 0
	for _, strategy := range request.Strategies {
		totalBaseHours += strategy.BaseRecoveryHours
	}
	for _, assoc := range associations {
		totalDataGB += assoc.DataToRecoverGB
	}

	var dataTransferHours float64 = 0
	if request.NetworkBandwidthMbps != nil && *request.NetworkBandwidthMbps > 0 {
		dataTransferHours = (float64(totalDataGB) * 8 * 1024) / (float64(*request.NetworkBandwidthMbps) * 3600)
	}

	skillMultiplier := 1.5
	if request.ItSkillLevel != nil {
		switch strings.ToLower(*request.ItSkillLevel) {
		case "средний":
			skillMultiplier = 1.0
		case "эксперт":
			skillMultiplier = 0.7
		}
	}

	docMultiplier := 1.5
	if request.DocumentationQuality != nil {
		switch strings.ToLower(*request.DocumentationQuality) {
		case "хорошая":
			docMultiplier = 1.0
		case "отличная":
			docMultiplier = 0.8
		}
	}

	finalTime := (totalBaseHours + dataTransferHours) * skillMultiplier * docMultiplier

	return r.db.Model(&request).Update("calculated_recovery_time_hours", finalTime).Error
}
