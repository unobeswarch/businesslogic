package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/unobeswarch/businesslogic/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsuarioExistente = errors.New("USER_ALREADY_EXISTS")
	ErrDatosEnviados    = errors.New("VALIDATION_ERROR")
	ErrTratamientoDatos = errors.New("BUSINESS_RULE_VIOLATION")
)

func RegistrarUsuario(u models.User) (int, time.Time, error) {
	if u.NombreCompleto == "" || u.Correo == "" || u.Contrasena == "" || len(u.Contrasena) < 8 {
		return 0, time.Time{}, ErrDatosEnviados
	}

	if !u.AceptaTratamientoDatos {
		return 0, time.Time{}, ErrTratamientoDatos
	}

	db, err := sql.Open("postgres", "postgres://postgres:123@localhost:5432/blogic_db")
	if err != nil {
		return 0, time.Time{}, err
	}
	defer db.Close()

	var existe bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE correo=$1 OR identificacion=$2)", u.Correo, u.Identificacion).Scan(&existe)
	if err != nil {
		return 0, time.Time{}, err
	}
	if existe {
		return 0, time.Time{}, ErrUsuarioExistente
	}

	hash_contrasena, err := bcrypt.GenerateFromPassword([]byte(u.Contrasena), bcrypt.DefaultCost)
	if err != nil {
		return 0, time.Time{}, err
	}

	query := `
		INSERT INTO usuarios 
		(nombre_completo, edad, rol, identificacion, correo, contrasena, acepta_tratamiento_datos)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, fecha_creacion
	`

	var id int
	var fechaCreacion time.Time

	err = db.QueryRow(
		query,
		u.NombreCompleto,
		u.Edad,
		u.Rol,
		u.Identificacion,
		u.Correo,
		hash_contrasena,
		u.AceptaTratamientoDatos,
	).Scan(&id, &fechaCreacion)
	if err != nil {
		return 0, time.Time{}, err
	}

	return id, fechaCreacion, nil
}

func IniciarSesion(correo string, contrasena string) string {

	return "login"
}

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
