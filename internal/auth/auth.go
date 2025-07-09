package auth

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	Token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	signedJWT, err := Token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedJWT, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	}, jwt.WithLeeway(5*time.Second))
	if err != nil {
		return uuid.UUID{}, err
	}
	id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(id)
}

func GetBearerToken(headers http.Header) (string, error) {
	bearing := headers.Get("Authorization")
	if bearing == "" {
		return bearing, fmt.Errorf("error: authorization header was empty: '%s'", bearing)
	}
	if len(bearing) < 8 {
		return bearing, fmt.Errorf("error: too short authorization string: '%s'", bearing)
	}
	if bearing[0:7] != "Bearer " {
		return bearing, fmt.Errorf("error: invalid authorization string: '%s'", bearing)
	}
	return bearing[7:], nil
}
