package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/manish-npx/go-echo-pg/internal/config"
	"github.com/manish-npx/go-echo-pg/internal/model"
)

type Claims struct {
	UserID pgtype.UUID `json:"user_id"`
	Email  string      `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(user *model.User, config *config.Config) (string, int64, error) {
	expirationTime := time.Now().Add(time.Duration(config.JWT.ExpiresIn) * time.Second)
	expiresAt := expirationTime.Unix()

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-echo-pg-app",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWT.Secret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt, nil
}

func ValidateToken(tokenString string, config *config.Config) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.JWT.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Additional validation
	if time.Until(claims.ExpiresAt.Time) < 0 {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}
