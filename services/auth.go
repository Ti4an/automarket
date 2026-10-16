package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"

	"automarket/models"
	"automarket/storage"
)

var (
	ErrInvalidInput       = errors.New("datos inválidos")
	ErrEmailTaken         = errors.New("el correo ya está registrado")
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrInvalidToken       = errors.New("token inválido o ausente")
)

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type AuthService struct {
	store    *storage.Store
	mu       sync.Mutex
	sessions map[string]string
}

func NewAuthService(store *storage.Store) *AuthService {
	return &AuthService{store: store, sessions: map[string]string{}}
}

func (a *AuthService) CreateUser(name, email, phone, password string, role models.Role) (models.User, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(email) == "" ||
		strings.TrimSpace(phone) == "" || len(password) < 8 {
		return models.User{}, ErrInvalidInput
	}
	if _, err := a.store.FindUserByEmail(email); err == nil {
		return models.User{}, ErrEmailTaken
	}
	id, err := randomHex(8)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{ID: id, Name: name, Email: email, Phone: phone, Role: role, Password: password}
	if err := a.store.AddUser(user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

// Login valida las credenciales y crea una sesión; devuelve el token y el usuario.
func (a *AuthService) Login(email, password string) (string, models.User, error) {
	user, err := a.store.FindUserByEmail(email)
	if err != nil || user.Password != password {
		return "", models.User{}, ErrInvalidCredentials
	}
	token, err := randomHex(32)
	if err != nil {
		return "", models.User{}, err
	}
	a.mu.Lock()
	a.sessions[token] = user.ID
	a.mu.Unlock()
	return token, user, nil
}

func (a *AuthService) Authenticate(token string) (models.User, error) {
	a.mu.Lock()
	userID, ok := a.sessions[token]
	a.mu.Unlock()
	if !ok {
		return models.User{}, ErrInvalidToken
	}
	user, err := a.store.FindUserByID(userID)
	if err != nil {
		return models.User{}, ErrInvalidToken
	}
	return user, nil
}