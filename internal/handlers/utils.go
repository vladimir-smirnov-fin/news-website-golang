package handlers

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SetFlashMessage(w http.ResponseWriter, msgType string, msg string) {
	value := msgType + "|" + msg
	cookie := http.Cookie{
		Name:     "flash_msg",
		Value:    value,
		Path:     "/",
		MaxAge:   5,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}

func GetFlashMessage(r *http.Request) (msgType, msg string) {
	cookie, err := r.Cookie("flash_msg")
	if err != nil {
		return "", ""
	}
	parts := strings.SplitN(cookie.Value, "|", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", cookie.Value
}

func ClearFlashMessage(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     "flash_msg",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}