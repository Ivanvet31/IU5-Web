package repository

import (
	"RIP/internal/app/ds"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) GetStrategies(title string) ([]ds.Strategy, error) {
	var strategies []ds.Strategy
	query := r.db.Where("status = ?", "active")
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if err := query.Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

func (r *Repository) GetStrategyByID(id uint) (ds.Strategy, error) {
	var strategy ds.Strategy
	if err := r.db.First(&strategy, id).Error; err != nil {
		return ds.Strategy{}, err
	}
	return strategy, nil
}

func (r *Repository) CreateStrategy(strategy *ds.Strategy) error {
	return r.db.Create(strategy).Error
}

func (r *Repository) UpdateStrategy(id uint, req ds.UpdateStrategyRequest) (*ds.Strategy, error) {
	var strategy ds.Strategy
	if err := r.db.First(&strategy, id).Error; err != nil {
		return nil, err
	}
	if req.Title != nil {
		strategy.Title = *req.Title
	}
	if req.Description != nil {
		strategy.Description = *req.Description
	}
	if req.BaseRecoveryHours != nil {
		strategy.BaseRecoveryHours = *req.BaseRecoveryHours
	}
	if err := r.db.Save(&strategy).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (r *Repository) DeleteStrategy(id uint) error {
	var strategy ds.Strategy
	var imageURLToDelete string

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&strategy, id).Error; err != nil {
			return err
		}
		if strategy.ImageURL != nil {
			imageURLToDelete = *strategy.ImageURL
		}
		if err := tx.Delete(&ds.Strategy{}, id).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	if imageURLToDelete != "" {
		parsedURL, err := url.Parse(imageURLToDelete)
		if err != nil {
			log.Printf("ERROR: could not parse image URL for deletion: %v", err)
			return nil
		}
		objectName := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", r.bucketName))
		err = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("ERROR: failed to delete object '%s' from MinIO: %v", objectName, err)
		}
	}
	return nil
}

func (r *Repository) UploadStrategyImage(strategyID uint, fileHeader *multipart.FileHeader) (string, error) {
	var finalImageURL string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var strategy ds.Strategy
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&strategy, strategyID).Error; err != nil {
			return fmt.Errorf("strategy with id %d not found: %w", strategyID, err)
		}
		if strategy.ImageURL != nil && *strategy.ImageURL != "" {
			oldImageURL, err := url.Parse(*strategy.ImageURL)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.bucketName))
				r.minioClient.RemoveObject(context.Background(), r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
			}
		}

		newUUID := uuid.New()
		ext := filepath.Ext(fileHeader.Filename)
		objectName := fmt.Sprintf("%s%s", newUUID, ext)

		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = r.minioClient.PutObject(context.Background(), r.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		})
		if err != nil {
			return fmt.Errorf("failed to upload to minio: %w", err)
		}
		imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, r.bucketName, objectName)
		if err := tx.Model(&strategy).Update("image_url", imageURL).Error; err != nil {
			return fmt.Errorf("failed to update strategy image url in db: %w", err)
		}
		finalImageURL = imageURL
		return nil
	})
	if err != nil {
		return "", err
	}
	return finalImageURL, nil
}
