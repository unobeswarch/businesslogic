package services

import (
	"database/sql"
	"time"

	"errors"

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
