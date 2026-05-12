package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"go-news-site-production/internal/models"
)

var sessions = struct {
	sync.RWMutex
	store map[string]models.Session
}{
	store: make(map[string]models.Session),
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateSession(userId int) (string, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}

	session := models.Session{
		UserId:    userId,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	sessions.Lock()
	sessions.store[token] = session
	sessions.Unlock()

	return token, nil
}

func GetUserIdFromSession(r *http.Request) (int, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, false
	}

	token := cookie.Value

	sessions.RLock()
	session, exists := sessions.store[token]
	sessions.RUnlock()

	if !exists {
		return 0, false
	}

	if time.Now().After(session.ExpiresAt) {
		sessions.Lock()
		delete(sessions.store, token)
		sessions.Unlock()
		return 0, false
	}

	return session.UserId, true
}

func DeleteSession(r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return
	}

	token := cookie.Value

	sessions.Lock()
	delete(sessions.store, token)
	sessions.Unlock()
}

func IsAuthenticated(r *http.Request) bool {
	_, ok := GetUserIdFromSession(r)
	return ok
}

func GetCurrentUserId(r *http.Request) (int, bool) {
	return GetUserIdFromSession(r)
}

func CanEditArticle(userId int, articleUserId int) bool {
	return userId == articleUserId
}

func CleanupExpiredSessions() {
	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		sessions.Lock()
		for token, session := range sessions.store {
			if time.Now().After(session.ExpiresAt) {
				delete(sessions.store, token)
			}
		}
		sessions.Unlock()
	}
}