package ds

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	IsModerator bool   `json:"is_moderator"`
}
