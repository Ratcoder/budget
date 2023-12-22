package server

import (
	"context"
	"net/http"
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"errors"
)

type Session struct {
	UserId  int
	Expires int
}

var sessions map[string]Session = make(map[string]Session)

func auth(r *http.Request) (int, error) {
	// Get session cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0, errors.New("Cookie not found")
	}

	// Get session
	session, ok := sessions[cookie.Value]
	if !ok {
		return 0, errors.New("Session not found")
	}

	// Check if session is expired
	if session.Expires < 0 {
		return 0, errors.New("Session expired")
	}

	// Check if session is about to expire
	if session.Expires < 60 {
		// Refresh session
		session.Expires = 60
		sessions[cookie.Value] = session
	}

	return session.UserId, nil
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, err := auth(r)
		if err != nil {
			http.Redirect(w, r, "/login.html", http.StatusSeeOther)
			return
		}

		// Call next handler
		r = r.WithContext(context.WithValue(r.Context(), "user", userId))
		next.ServeHTTP(w, r)
	})
}

func createSession(userId int) (string, error) {
	// Create session
	session := Session{
		UserId:  userId,
		Expires: 0,
	}
	sessionId := make([]byte, 32)
	_, err := rand.Read(sessionId)
	if err != nil {
		return "", err
	}
	sessionString := base64.StdEncoding.EncodeToString(sessionId)
	sessions[sessionString] = session

	return sessionString, nil
}

func loginUser(username string, password string) (string, error) {
	// Get user from database
	dbUser, err := (*db).GetUserByName(username)
	if err != nil {
		return "", err
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password))
	if err != nil {
		return "", err
	}

	// Create session
	sessionString, err := createSession(dbUser.Id)
	if err != nil {
		return "", err
	}

	return sessionString, nil
}