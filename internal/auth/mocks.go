package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TestAuthenticator struct {
}

func NewMockAuthenticator() *TestAuthenticator {
	return &TestAuthenticator{}
}

var testClaims = jwt.MapClaims{
	"aud": "test-aud",
	"iss": "test-issuer",
	"sub": int64(13),
	"exp": time.Now().Add(time.Hour).Unix(),
}

func (a *TestAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)
	return token.SignedString([]byte("secret"))
}

func (a *TestAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
}
