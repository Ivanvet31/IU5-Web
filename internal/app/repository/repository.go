package repository

import (
	"fmt"
	"strings"
)

type Strategy struct {
	ID          int
	Title       string
	ImageURL    string
	Description string
}

type Repository struct {
	strategies []Strategy
}

func NewRepository() *Repository {
	minioBaseURL := "http://localhost:9000/recovery-images"
	strategies := []Strategy{
		{
			ID:          1,
			Title:       "Backup & Recovery",
			ImageURL:    fmt.Sprintf("%s/backup.png", minioBaseURL),
			Description: "Самая распространенная стратегия. Предлагает регулярное создание копий данных и их сохранение на другом носителе или в другом месте. В случае сбоя данные восстанавливаются из резервной копии.",
		},
		{
			ID:          2,
			Title:       "Репликация",
			ImageURL:    fmt.Sprintf("%s/replication.png", minioBaseURL),
			Description: "Репликация данных — это процесс создания и поддержания нескольких копий одних и тех же данных на разных серверах или в разных хранилищах для обеспечения их доступности и отказоустойчивости.",
		},
		{
			ID:          3,
			Title:       "Горячий резерв",
			ImageURL:    fmt.Sprintf("%s/hot_reserve.png", minioBaseURL),
			Description: "Горячий резерв предполагает наличие полностью дублирующей инфраструктуры, которая работает параллельно с основной и готова немедленно принять на себя всю нагрузку в случае сбоя.",
		},
		{
			ID:          4,
			Title:       "Теплый резерв",
			ImageURL:    fmt.Sprintf("%s/warm_reserve.png", minioBaseURL),
			Description: "Теплый резерв — это компромисс между горячим и холодным резервом. Инфраструктура частично запущена и регулярно обновляется, что позволяет сократить время восстановления по сравнению с холодным резервом.",
		},
		{
			ID:          5,
			Title:       "Холодный резерв",
			ImageURL:    fmt.Sprintf("%s/cold_reserve.png", minioBaseURL),
			Description: "Холодный резерв — это наличие резервной инфраструктуры, которая выключена и требует ручного запуска и настройки в случае аварии. Это самый дешевый, но и самый медленный способ восстановления.",
		},
	}
	return &Repository{strategies: strategies}
}

func (r *Repository) GetStrategies(query string) ([]Strategy, error) {
	// Если строка поиска пустая, возвращаем все стратегии
	if query == "" {
		return r.strategies, nil
	}

	var filteredStrategies []Strategy
	lowerQuery := strings.ToLower(query) // Приводим запрос к нижнему регистру

	for _, s := range r.strategies {
		// Если заголовок (в нижнем регистре) содержит подстроку из запроса
		if strings.Contains(strings.ToLower(s.Title), lowerQuery) {
			filteredStrategies = append(filteredStrategies, s)
		}
	}

	return filteredStrategies, nil
}

func (r *Repository) GetStrategyByID(id int) (Strategy, error) {
	for _, s := range r.strategies {
		if s.ID == id {
			return s, nil
		}
	}
	return Strategy{}, fmt.Errorf("стратегия с ID %d не найдена", id)
}
