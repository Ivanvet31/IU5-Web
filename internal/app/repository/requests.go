package repository

import (
	"RIP/internal/app/ds"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GetOrCreateDraftRequest находит или создает черновик заявки
func (r *Repository) GetOrCreateDraftRequest(userID uint) (ds.Request, error) {
	var request ds.Request
	err := r.db.Where(ds.Request{UserID: userID, Status: "draft"}).Attrs(ds.Request{CreatedAt: time.Now()}).FirstOrCreate(&request).Error
	return request, err
}

// AddStrategyToRequest добавляет стратегию в заявку
func (r *Repository) AddStrategyToRequest(requestID, strategyID uint) error {
	var existing ds.RequestStrategy
	err := r.db.Where("request_id = ? AND strategy_id = ?", requestID, strategyID).First(&existing).Error
	if err == nil {
		return nil // Уже существует
	}
	if err != gorm.ErrRecordNotFound {
		return err // Другая ошибка БД
	}

	association := ds.RequestStrategy{
		RequestID:       requestID,
		StrategyID:      strategyID,
		DataToRecoverGB: 0,
	}
	return r.db.Create(&association).Error
}

// GetRequestWithStrategies загружает заявку со связанными стратегиями, пользователями
func (r *Repository) GetRequestWithStrategies(requestID uint) (ds.Request, error) {
	var request ds.Request
	err := r.db.Preload("Strategies").Preload("User").Preload("Moderator").First(&request, requestID).Error
	if err == nil && request.Status == "deleted" {
		return ds.Request{}, errors.New("request has been deleted")
	}
	return request, err
}

// LogicallyDeleteRequest удаляет заявку (только создатель, только черновик)
func (r *Repository) LogicallyDeleteRequest(requestID uint, creatorID uint) error {
	var request ds.Request
	if err := r.db.First(&request, requestID).Error; err != nil {
		return err
	}
	if request.UserID != creatorID {
		return errors.New("only creator can delete the request")
	}
	if request.Status != "draft" {
		return errors.New("only draft request can be deleted")
	}
	return r.db.Model(&request).Update("status", "deleted").Error
}

// UpdateRequestDetails обновляет поля заявки, введенные пользователем
func (r *Repository) UpdateRequestDetails(requestID uint, req ds.UpdateRequestDetailsRequest) error {
	updates := make(map[string]interface{})
	if req.ItSkillLevel != nil {
		updates["it_skill_level"] = *req.ItSkillLevel
	}
	if req.NetworkBandwidthMbps != nil {
		updates["network_bandwidth_mbps"] = *req.NetworkBandwidthMbps
	}
	if req.DocumentationQuality != nil {
		updates["documentation_quality"] = *req.DocumentationQuality
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.Model(&ds.Request{}).Where("id = ?", requestID).Updates(updates).Error
}

// UpdateRequestStrategyData обновляет поле в связующей таблице
func (r *Repository) UpdateRequestStrategyData(requestID, strategyID uint, req ds.UpdateRequestStrategyRequest) error {
	return r.db.Model(&ds.RequestStrategy{}).
		Where("request_id = ? AND strategy_id = ?", requestID, strategyID).
		Update("data_to_recover_gb", req.DataToRecoverGB).Error
}

// ListRequestsFiltered возвращает список заявок с фильтрами
func (r *Repository) ListRequestsFiltered(status, from, to string) ([]ds.Request, error) {
	var requests []ds.Request
	query := r.db.Preload("User").Preload("Moderator")
	query = query.Where("status != ? AND status != ?", "draft", "deleted")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if from != "" {
		if fromDate, err := time.Parse("2006-01-02", from); err == nil {
			query = query.Where("formed_at >= ?", fromDate)
		}
	}
	if to != "" {
		if toDate, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where("formed_at <= ?", toDate)
		}
	}

	if err := query.Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// FormRequest меняет статус с 'draft' на 'formed'
func (r *Repository) FormRequest(requestID, creatorID uint) error {
	var request ds.Request
	if err := r.db.First(&request, requestID).Error; err != nil {
		return err
	}
	if request.UserID != creatorID {
		return errors.New("only creator can form the request")
	}
	if request.Status != "draft" {
		return errors.New("only draft request can be formed")
	}
	return r.db.Model(&request).Updates(map[string]interface{}{"status": "formed", "formed_at": time.Now()}).Error
}

// ResolveRequest меняет статус с 'formed' на 'completed' или 'rejected'
func (r *Repository) ResolveRequest(requestID, moderatorID uint, action string) error {
	var request ds.Request
	if err := r.db.First(&request, requestID).Error; err != nil {
		return err
	}
	if request.Status != "formed" {
		return errors.New("only formed request can be resolved")
	}

	updates := map[string]interface{}{"moderator_id": &moderatorID, "completed_at": time.Now()}
	switch action {
	case "complete":
		updates["status"] = "completed"
		if err := r.CalculateAndSaveRecoveryTime(requestID); err != nil {
			return err
		}
	case "reject":
		updates["status"] = "rejected"
	default:
		return errors.New("invalid action: must be 'complete' or 'reject'")
	}
	return r.db.Model(&request).Updates(updates).Error
}

// CalculateAndSaveRecoveryTime рассчитывает и сохраняет итоговое время
func (r *Repository) CalculateAndSaveRecoveryTime(requestID uint) error {
	var request ds.Request
	if err := r.db.Preload("Strategies").First(&request, requestID).Error; err != nil {
		return err
	}

	var associations []ds.RequestStrategy
	if err := r.db.Where("request_id = ?", requestID).Find(&associations).Error; err != nil {
		return err
	}

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

// RemoveStrategyFromRequest удаляет связь м-м
func (r *Repository) RemoveStrategyFromRequest(requestID, strategyID uint) error {
	result := r.db.Where("request_id = ? AND strategy_id = ?", requestID, strategyID).Delete(&ds.RequestStrategy{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("association not found")
	}
	return nil
}
