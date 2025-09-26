package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type AuthService struct {
	// En un escenario real, aquí tendrías configuración para validar JWT
	// Por ahora implementamos una validación básica
}

type UserClaims struct {
	UserID string
	Email  string
	Role   string
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// ValidateTokenAndRole valida el token de autorización y verifica el rol
func (s *AuthService) ValidateTokenAndRole(ctx context.Context, authHeader string, requiredRole string) (*UserClaims, error) {
	if authHeader == "" {
		return nil, errors.New("token de autorización requerido")
	}

	// Extraer el token del header "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("formato de token inválido")
	}

	token := parts[1]
	
	// En un escenario real, aquí validarías el JWT contra tu servicio de auth
	// Por ahora implementamos una validación mock
	userClaims, err := s.validateJWT(token)
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	// Verificar rol
	if userClaims.Role != requiredRole {
		return nil, fmt.Errorf("acceso denegado: se requiere rol %s, pero el usuario tiene rol %s", 
			requiredRole, userClaims.Role)
	}

	return userClaims, nil
}

// validateJWT - Mock implementation. En producción usarías una librería JWT real
func (s *AuthService) validateJWT(token string) (*UserClaims, error) {
	// Mock validation - en producción aquí validarías el token real
	// y extraerías los claims del JWT
	
	// Para testing, aceptamos tokens con formato específico
	switch token {
	case "doctor_token_123":
		return &UserClaims{
			UserID: "doctor_1",
			Email:  "doctor@hospital.com",
			Role:   "doctor",
		}, nil
	case "patient_token_456":
		return &UserClaims{
			UserID: "patient_1", 
			Email:  "patient@email.com",
			Role:   "patient",
		}, nil
	default:
		return nil, errors.New("token no válido")
	}
}