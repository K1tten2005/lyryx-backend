package auth

import "github.com/golang-jwt/jwt/v5"

type JwtCustomClaims struct {
	UserID     int    `json:"user_id"`
	Email      string `json:"email"`
	Role       string `json:"role,omitempty"`
	jwt.RegisteredClaims
}
