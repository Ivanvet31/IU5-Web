package repository

import (
	"RIP/internal/app/ds"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// GetOrCreateDraftRequest находит или создает черновик заявки
func (r *Repository) GetOrCreateDraftRequest(userID uint) (ds.Recovery_request, error) {
	var request ds.Recovery_request
	err := r.db.Where(ds.Recovery_request{UserID: userID, Status: "draft"}).Attrs(ds.Recovery_request{CreatedAt: time.Now()}).FirstOrCreate(&request).Error
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
func (r *Repository) GetRequestWithStrategies(requestID uint) (ds.Recovery_request, error) {
	var request ds.Recovery_request
	err := r.db.Preload("Strategies").Preload("User").Preload("Moderator").First(&request, requestID).Error
	if err == nil && request.Status == "deleted" {
		return ds.Recovery_request{}, errors.New("request has been deleted")
	}
	return request, err
}

// LogicallyDeleteRequest удаляет заявку (только создатель, только черновик)
func (r *Repository) LogicallyDeleteRequest(requestID uint, creatorID uint) error {
	var request ds.Recovery_request
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
	return r.db.Model(&ds.Recovery_request{}).Where("id = ?", requestID).Updates(updates).Error
}

// UpdateRequestStrategyData обновляет поле в связующей таблице
func (r *Repository) UpdateRequestStrategyData(requestID, strategyID uint, req ds.UpdateRequestStrategyRequest) error {
	return r.db.Model(&ds.RequestStrategy{}).
		Where("request_id = ? AND strategy_id = ?", requestID, strategyID).
		Update("data_to_recover_gb", req.DataToRecoverGB).Error
}

// ListRequestsFiltered возвращает список заявок с фильтрами
func (r *Repository) ListRequestsFiltered(userID uint, isModerator bool, status, from, to string) ([]ds.Recovery_request, error) {
	var requests []ds.Recovery_request
	query := r.db.Preload("User").Preload("Moderator")
	query = query.Where("status != ? AND status != ?", "draft", "deleted")

	if !isModerator {
		query = query.Where("user_id = ?", userID)
	}

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
	var request ds.Recovery_request
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
// И отправляет задачу в Python сервис при complete
func (r *Repository) ResolveRequest(requestID, moderatorID uint, action string) error {
	var request ds.Recovery_request
	// Загружаем заявку со стратегиями для формирования payload
	if err := r.db.Preload("Strategies").First(&request, requestID).Error; err != nil {
		return err
	}
	if request.Status != "formed" {
		return errors.New("only formed request can be resolved")
	}

	updates := map[string]interface{}{"moderator_id": &moderatorID, "completed_at": time.Now()}
	
	if action == "reject" {
		updates["status"] = "rejected"
		return r.db.Model(&request).Updates(updates).Error
	} else if action == "complete" {
		updates["status"] = "completed"
		updates["calculated_recovery_time_hours"] = nil // Сбрасываем, пока не посчитает Python

		// Сохраняем статус
		if err := r.db.Model(&request).Updates(updates).Error; err != nil {
			return err
		}

		// --- Подготовка данных для асинхронного сервиса ---
		var strategiesData []ds.AsyncStrategyData
		for _, strategy := range request.Strategies {
			var assoc ds.RequestStrategy
			// Получаем данные из M-M (объем данных)
			r.db.Where("request_id = ? AND strategy_id = ?", requestID, strategy.ID).First(&assoc)

			strategiesData = append(strategiesData, ds.AsyncStrategyData{
				BaseRecoveryHours: strategy.BaseRecoveryHours,
				DataToRecoverGB:   assoc.DataToRecoverGB,
			})
		}

		payload := ds.AsyncCalculateRequest{
			ID:                   request.ID,
			ItSkillLevel:         "",
			NetworkBandwidthMbps: 0,
			DocumentationQuality: "",
			Strategies:           strategiesData,
		}
		if request.ItSkillLevel != nil {
			payload.ItSkillLevel = *request.ItSkillLevel
		}
		if request.NetworkBandwidthMbps != nil {
			payload.NetworkBandwidthMbps = *request.NetworkBandwidthMbps
		}
		if request.DocumentationQuality != nil {
			payload.DocumentationQuality = *request.DocumentationQuality
		}

		// --- Асинхронная отправка запроса ---
		go func() {
			jsonData, _ := json.Marshal(payload)
			// Адрес Python сервиса на порту 8001
			resp, err := http.Post("http://localhost:8001/calculate", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("ERROR: Failed to send task to async service: %v\n", err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("Task sent to async service. Status: %s\n", resp.Status)
		}()

		return nil
	} else {
		return errors.New("invalid action: must be 'complete' or 'reject'")
	}
}

// UpdateRecoveryTime обновляет рассчитанное время (вызывается асинхронным сервисом)
func (r *Repository) UpdateRecoveryTime(requestID uint, timeVal float64) error {
	return r.db.Model(&ds.Recovery_request{}).
		Where("id = ?", requestID).
		Update("calculated_recovery_time_hours", timeVal).Error
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
