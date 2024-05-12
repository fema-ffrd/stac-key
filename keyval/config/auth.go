package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func Ping(cfg *APIConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Pong")
	}
}

func Validate(cfg *APIConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("X-Forwarded-User", "stac-user")
		return c.JSON(http.StatusOK, "Validated")
	}
}

type Claims struct {
	UserName    string              `json:"preferred_username"`
	Email       string              `json:"email"`
	RealmAccess map[string][]string `json:"realm_access"`
	jwt.RegisteredClaims
}

type PublicKey struct {
	Kid  string   `json:"kid"`
	Kty  string   `json:"kty"`
	Alg  string   `json:"alg"`
	Use  string   `json:"use"`
	N    string   `json:"n"`
	E    string   `json:"e"`
	X5C  []string `json:"x5c"`
	X5T  string   `json:"x5t"`
	S256 string   `json:"x5t#S256"`
}

var publicKeys []PublicKey

func getPublicKeys() ([]PublicKey, error) {
	log.Info(os.Getenv("KEYCLOAK_PUBLIC_KEYS_URL"))
	r, err := http.Get(os.Getenv("KEYCLOAK_PUBLIC_KEYS_URL"))
	if err != nil {
		return []PublicKey{}, err
	}

	defer r.Body.Close()

	var target map[string][]PublicKey

	if err = json.NewDecoder(r.Body).Decode(&target); err != nil {
		return []PublicKey{}, err
	}

	return target["keys"], nil
}

func init() {
	var err error
	publicKeys, err = getPublicKeys()
	if err != nil {
		panic(err)
	}
}

func getPublicKeyStr(kid string) string {
	var publicKeyStr string
	for _, key := range publicKeys {
		if key.Kid == kid {
			publicKeyStr = "-----BEGIN CERTIFICATE-----\n" + key.X5C[0] + "\n-----END CERTIFICATE-----"
		}
	}
	return publicKeyStr
}

func validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		publicKeyStr := getPublicKeyStr(token.Header["kid"].(string))

		if publicKeyStr == "" {
			var err error
			publicKeys, err = getPublicKeys()
			if err != nil {
				return nil, err
			}
			publicKeyStr = getPublicKeyStr(token.Header["kid"].(string))
		}

		return jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyStr))
	})

	if err != nil {
		// Provide more context on the error when parsing the token
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, fmt.Errorf("failed to parse JWT: malformed token")
		} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			// Invalid signature
			return nil, fmt.Errorf("failed to parse JWT: invalid signature")
		} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			// Token is either expired or not active yet
			return nil, fmt.Errorf("failed to parse JWT: token expired")
		}
		return nil, fmt.Errorf("failed to parse JWT: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	return claims, nil
}

func overlap(s1 []string, s2 []string) bool {
	for _, x := range s1 {
		for _, y := range s2 {
			if x == y {
				return true
			}
		}
	}
	return false
}

func Authorize(handler echo.HandlerFunc, allowedRoles ...string) echo.HandlerFunc {
	return func(c echo.Context) error {

		headers := c.Request().Header

		authHead := headers.Get("Authorization")

		// Check if the Authorization header is missing or not in the expected format
		if authHead == "" || !strings.HasPrefix(authHead, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, "missing or invalid authorization header")
		}

		tokenString := strings.Split(authHead, "Bearer ")[1]

		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, "missing authorization header")
		}

		claims, err := validateToken(tokenString)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}

		// Store the claims in the echo.Context
		c.Set("claims", claims)

		ok := overlap(claims.RealmAccess["roles"], allowedRoles)
		if !ok {
			return c.JSON(http.StatusUnauthorized, "user is not authorized")
		}
		return handler(c)
	}
}
