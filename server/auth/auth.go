package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

var (
	keys     *jwks
	keysOnce sync.Once
	keysErr  error
)

// Load JWKS once (safe + cacheable)
func loadJWKS() {
	projectURL := os.Getenv("SUPABASE_URL")
	if projectURL == "" {
		keysErr = errors.New("SUPABASE_URL not set")
		return
	}

	resp, err := http.Get(projectURL + "/auth/v1/keys")
	if err != nil {
		keysErr = err
		return
	}
	defer resp.Body.Close()

	var k jwks
	if err := json.NewDecoder(resp.Body).Decode(&k); err != nil {
		keysErr = err
		return
	}

	keys = &k
}

func getPublicKey(kid string) (*rsa.PublicKey, error) {
	for _, k := range keys.Keys {
		if k.Kid == kid && k.Kty == "RSA" {
			nBytes, _ := base64.RawURLEncoding.DecodeString(k.N)
			eBytes, _ := base64.RawURLEncoding.DecodeString(k.E)

			n := new(big.Int).SetBytes(nBytes)
			e := new(big.Int).SetBytes(eBytes).Int64()

			return &rsa.PublicKey{
				N: n,
				E: int(e),
			}, nil
		}
	}
	return nil, errors.New("signing key not found")
}

// VerifySupabaseJWT validates the token and returns the user ID
func VerifySupabaseJWT(tokenString string) (string, error) {
	keysOnce.Do(loadJWKS)
	if keysErr != nil {
		return "", keysErr
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected signing algorithm")
		}

		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid")
		}

		return getPublicKey(kid)
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("missing sub claim")
	}

	return sub, nil
}

// HTTP middleware
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if len(h) < 8 || h[:7] != "Bearer " {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := VerifySupabaseJWT(h[7:])
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

