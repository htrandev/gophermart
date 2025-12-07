package authorizer

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type claims struct {
	jwt.RegisteredClaims
	UserId string `json:"Id"`
}

type Authorizer struct {
	secretKey []byte
	ttl       time.Duration
}

func New(key string, ttl time.Duration) *Authorizer {
	return &Authorizer{
		secretKey: []byte(key),
		ttl:       ttl,
	}
}

func (a *Authorizer) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("generate hash from password: %w", err)
	}
	return string(hash), nil
}

func (a *Authorizer) ValidatePassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil

}

func (a *Authorizer) Token(id string) (string, error) {
	ttl := time.Now().Add(a.ttl)
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(ttl),
		},
		UserId: id,
	})

	token, err := t.SignedString(a.secretKey)
	if err != nil {
		return "", fmt.Errorf("signed string: %w", err)
	}
	return token, nil
}

func (a *Authorizer) GetIdFromToken(token string) (string, error) {
	parseen, err := jwt.ParseWithClaims(
		token,
		&claims{},
		func(token *jwt.Token) (any, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, ErrUnexpectedMethod
			}

			return a.secretKey, nil
		},
	)
	if err != nil {
		return "", err
	}

	claims, ok := parseen.Claims.(*claims)
	if !ok {
		return "", ErrInvalidClaims
	}

	return claims.UserId, nil
}
