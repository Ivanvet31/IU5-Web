// internal/app/models/models.go
package models

import "time"

// User соответствует таблице 'users'
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"size:150;unique;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	Email        string `gorm:"size:254"`
	IsActive     bool   `gorm:"not null;default:true"`
	IsModerator  bool   `gorm:"not null;default:false"`
}

// Strategy соответствует таблице 'strategies'
type Strategy struct {
	ID                uint    `gorm:"primaryKey"`
	Title             string  `gorm:"size:255;not null"`
	Description       string  `gorm:"type:text"`
	ImageURL          *string `gorm:"size:2048"` // Указатель *string делает поле nullable
	Status            string  `gorm:"size:50;not null;default:'active'"`
	BaseRecoveryHours float64
}

// Request соответствует таблице 'requests'
type Request struct {
	ID                          uint       `gorm:"primaryKey"`
	Status                      string     `gorm:"size:50;not null;default:'draft'"`
	CreatedAt                   time.Time  `gorm:"not null"`
	FormedAt                    *time.Time // Указатель *time.Time делает поле nullable
	CompletedAt                 *time.Time
	UserID                      uint    `gorm:"not null"`
	User                        User    `gorm:"foreignKey:UserID"`
	ModeratorID                 *uint   // Указатель *uint делает поле nullable
	Moderator                   *User   `gorm:"foreignKey:ModeratorID"`
	ItSkillLevel                *string `gorm:"size:50"`
	NetworkBandwidthMbps        *int
	DocumentationQuality        *string `gorm:"size:50"`
	CalculatedRecoveryTimeHours *float64
	Strategies                  []Strategy `gorm:"many2many:request_strategies;"` // Связь многие-ко-многим
}

// RequestStrategy соответствует связующей таблице 'request_strategies'
type RequestStrategy struct {
	RequestID       uint `gorm:"primaryKey"`
	StrategyID      uint `gorm:"primaryKey"`
	DataToRecoverGB int  `gorm:"not null"`
}
