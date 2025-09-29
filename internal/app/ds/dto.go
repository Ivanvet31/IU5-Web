package ds

import "time"

// --- Strategy DTOs ---

type StrategyDTO struct {
	ID                uint    `json:"id"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	ImageURL          *string `json:"image_url"`
	BaseRecoveryHours float64 `json:"base_recovery_hours"`
}

type CreateStrategyRequest struct {
	Title             string  `json:"title" binding:"required"`
	Description       string  `json:"description" binding:"required"`
	BaseRecoveryHours float64 `json:"base_recovery_hours" binding:"required"`
}

type UpdateStrategyRequest struct {
	Title             *string  `json:"title"`
	Description       *string  `json:"description"`
	BaseRecoveryHours *float64 `json:"base_recovery_hours"`
}

// --- Request DTOs ---

type CartBadgeDTO struct {
	RequestID *uint `json:"request_id"`
	Count     int   `json:"count"`
}

type RequestDTO struct {
	ID                          uint          `json:"id"`
	Status                      string        `json:"status"`
	CreatedAt                   time.Time     `json:"created_at"`
	CreatorUsername             string        `json:"creator_username"`
	ModeratorUsername           *string       `json:"moderator_username,omitempty"`
	ItSkillLevel                *string       `json:"it_skill_level,omitempty"`
	NetworkBandwidthMbps        *int          `json:"network_bandwidth_mbps,omitempty"`
	DocumentationQuality        *string       `json:"documentation_quality,omitempty"`
	CalculatedRecoveryTimeHours *float64      `json:"calculated_recovery_time_hours,omitempty"`
	Strategies                  []StrategyDTO `json:"strategies,omitempty"`
}

type UpdateRequestDetailsRequest struct {
	ItSkillLevel         *string `json:"it_skill_level"`
	NetworkBandwidthMbps *int    `json:"network_bandwidth_mbps"`
	DocumentationQuality *string `json:"documentation_quality"`
}

type ResolveRequest struct {
	Action string `json:"action" binding:"required"` // "complete" or "reject"
}

// --- M-M DTOs ---

type UpdateRequestStrategyRequest struct {
	DataToRecoverGB int `json:"data_to_recover_gb" binding:"required"`
}

// --- User DTOs ---

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserDTO struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	IsModerator bool   `json:"is_moderator"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type UpdateUserRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}
