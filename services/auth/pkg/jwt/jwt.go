package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenProvider struct {
	SigningKey string
	SigningAlg string
}

func NewTokenProvider(signingKey string, signingAlg string) *TokenProvider {
	return &TokenProvider{signingKey, signingAlg}
}

func (tp *TokenProvider) NewToken(expires time.Duration, claims map[string]any) (string, error) {
	if expires <= 0 {
		panic("expires must be greater than 0")
	}
	claims["exp"] = time.Now().Add(expires).Unix()
	// add issued at claim to ensure token uniqueness
	claims["iat"] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.GetSigningMethod(tp.SigningAlg), jwt.MapClaims(claims))
	return token.SignedString([]byte(tp.SigningKey))
}

func (tp *TokenProvider) ParseClaimsFromToken(token string) (map[string]any, error) {
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(tp.SigningKey), nil
	}, jwt.WithValidMethods([]string{tp.SigningAlg}))
	if err != nil {
		return nil, err
	}
	return map[string]any(parsed.Claims.(jwt.MapClaims)), nil
}
