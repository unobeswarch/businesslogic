package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

var (
	ErrUsuarioExistente = errors.New("USER_ALREADY_EXISTS")
	ErrDatosEnviados    = errors.New("VALIDATION_ERROR")
	ErrTratamientoDatos = errors.New("BUSINESS_RULE_VIOLATION")
)

type AuthService struct {
	key []byte
}

type UserClaims struct {
	UserID string
	Email  string
	Role   string
	Name   string
}

type AuthResponse struct {
	Valid bool        `json:"valid"`
	User  interface{} `json:"user"`
	Error string      `json:"error,omitempty"`
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// UserExists verifica si el usuario existe en la base de datos relacional
func (s *AuthService) UserExistsAuthBE(ctx context.Context, userID string) (bool, error) {
	// Intentar primera con connection string sin URL encoding
	url := "http://localhost:8081/userExists"
	// Si falla, probar con URL encoding (línea comentada abajo)
	// db, err := sql.Open("postgres", "postgres://postgres:BDatosPost0912%2B@localhost:5432/blogic_db?sslmode=disable")
	// Original password (comentada):
	// db, err := sql.Open("postgres", "postgres://postgres:123@localhost:5432/blogic_db?sslmode=disable")

	bodyData, err := json.Marshal(map[string]string{
		"user_id": userID,
	})
	if err != nil {
		return false, fmt.Errorf("error al crear body JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyData))
	if err != nil {
		return false, fmt.Errorf("error al crear request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error al conectar con auth-be: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var msg map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&msg)
		return false, fmt.Errorf("auth-be respondió con %d: %v", resp.StatusCode, msg)
	}

	var res struct {
		Exists bool `json:"exists"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	return res.Exists, nil
}

func (s *AuthService) ValidateTokenWithAuthBE(ctx context.Context, authHeader string, requiredRole string) (*UserClaims, error) {
	body, _ := json.Marshal(map[string]string{"required_role": requiredRole})

	req, _ := http.NewRequestWithContext(ctx, "POST", "http://localhost:8081/validation", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error comunicando con auth-be: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result struct {
			UserID string `json:"UserID"`
			Email  string `json:"Email"`
			Role   string `json:"Role"`
			Name   string `json:"Name"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("error parseando respuesta: %w", err)
		}

		return &UserClaims{
			UserID: result.UserID,
			Email:  result.Email,
			Role:   result.Role,
			Name:   result.Name,
		}, nil

	case http.StatusUnauthorized:
		return nil, fmt.Errorf("token inválido o no proporcionado")

	case http.StatusForbidden:
		return nil, fmt.Errorf("rol no autorizado para esta operación (se requiere '%s')", requiredRole)

	default:
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		msg := errResp["error"]
		if msg == "" {
			msg = fmt.Sprintf("auth-be respondió con código %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("validación fallida: %s", msg)
	}
}
