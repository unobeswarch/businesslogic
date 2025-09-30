package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/unobeswarch/businesslogic/internal/models"
	"github.com/unobeswarch/businesslogic/internal/services"
)

func HandlerRegistrarUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Metodo no permitido",
		})
		return
	}

	var usuario models.User
	err := json.NewDecoder(r.Body).Decode(&usuario)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Problema durante la conversion a JSON",
		})
		return
	}

	id, fecha, err := services.RegistrarUsuario(usuario)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch err {
		case services.ErrDatosEnviados:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "VALIDATION_ERROR",
				"mensaje": "Datos de entrada inválidos",
			})
		case services.ErrUsuarioExistente:
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "USER_ALREADY_EXISTS",
				"mensaje": "Ya existe un usuario con este correo o identificación",
			})
		case services.ErrTratamientoDatos:
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "BUSINESS_RULE_VIOLATION",
				"mensaje": "Debe aceptar el tratamiento de datos personales",
			})
		default:
			// Log del error específico para depuración
			fmt.Printf("Error específico durante registro: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":         "INTERNAL_ERROR",
				"mensaje":       "Error interno del servidor",
				"codigo_error":  "REG_001",
				"error_detalle": err.Error(),
			})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":             id,
		"mensaje":        "Usuario registrado exitosamente",
		"fecha_registro": fecha,
	})
}

func HandlerIniciarSesion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Metodo no permitido",
		})
		return
	}

	var datos map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&datos)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Problema durante la conversion del JSON",
		})
		return
	}

	nombre, id, rol, token, err := services.IniciarSesion(datos["correo"].(string), datos["contrasena"].(string))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nombre":  nombre,
		"token":   token,
		"rol":     rol,
		"user_id": id,
		"correo":  datos["correo"].(string),
	})
}

func HandlerValidacion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Metodo no permitido",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error": "token de autorización requerido"}`, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, `{"error": "formato de token inválido"}`, http.StatusUnauthorized)
		return
	}
	token := parts[1]

	authService := services.NewAuthService()
	claims, err := authService.ValidateJWT(token)
	if err != nil {
		http.Error(w, `{"error": "token inválido"}`, http.StatusUnauthorized)
		return
	}

	resp := services.UserClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		Name:   claims.Name,
	}

	json.NewEncoder(w).Encode(resp)
}
