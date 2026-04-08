package services

import (
	"errors"
	"time"

	"xykitchen/backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Realm    string `json:"realm,omitempty"`
	jwt.RegisteredClaims
}

func ParseJWT(cfg *config.Config, tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := tok.Claims.(*Claims); ok && tok.Valid {
		return c, nil
	}
	return nil, errors.New("invalid token")
}

func SignJWT(cfg *config.Config, id int, username, role, realm string) (string, error) {
	dur, err := time.ParseDuration(cfg.JWT.ExpiresIn)
	if err != nil {
		dur = 168 * time.Hour
	}
	claims := Claims{
		ID:       id,
		Username: username,
		Role:     role,
		Realm:    realm,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(dur)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(cfg.JWT.Secret))
}
