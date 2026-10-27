package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 8 * time.Hour

type signinReq struct {
	Password string `json:"password"`
}

// passwordHash — полезная нагрузка токена: по ней нельзя восстановить пароль,
// а при смене TODO_PASSWORD старые токены перестают проходить проверку.
func passwordHash(pass string) string {
	sum := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sum[:])
}

func signKey(pass string) []byte {
	return []byte("go_final_project:" + pass)
}

func createToken(pass string) (string, error) {
	claims := jwt.MapClaims{
		"hash": passwordHash(pass),
		"exp":  time.Now().Add(tokenTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(signKey(pass))
}

func validToken(tokenString, pass string) bool {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return signKey(pass), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	hash, _ := claims["hash"].(string)
	return hash == passwordHash(pass)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("метод не поддерживается"))
		return
	}

	pass := os.Getenv("TODO_PASSWORD")
	if pass == "" {
		writeError(w, http.StatusBadRequest, errors.New("пароль не задан на сервере"))
		return
	}

	var req signinReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Password != pass {
		writeError(w, http.StatusUnauthorized, errors.New("неверный пароль"))
		return
	}

	token, err := createToken(pass)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// auth пропускает запрос дальше, только если пароль не задан
// или в куке token лежит действительный токен для текущего пароля.
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if pass != "" {
			var token string
			if cookie, err := r.Cookie("token"); err == nil {
				token = cookie.Value
			}
			if !validToken(token, pass) {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
