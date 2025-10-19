package services

import (
	"fmt"

	"github.com/unobeswarch/businesslogic/internal/clients"
	"github.com/unobeswarch/businesslogic/internal/models"
)

type DiagnosticService struct {
	client             *clients.PreDiagnosticClient
	notificationClient *clients.NotificationClient
	authClient         *clients.AuthClient
}

func NewDiagnosticService(baseURL, notificationURL, authURL string) *DiagnosticService {
	return &DiagnosticService{
		client:             clients.NewPrediagnosticClient(baseURL),
		notificationClient: clients.NewNotificationClient(notificationURL),
		authClient:         clients.NewAuthClient(authURL),
	}
}

// CreateDiagnostic procesa la creación de un diagnóstico
func (s *DiagnosticService) CreateDiagnostic(prediagnosticID, aprobacion, comentario string) (*models.DiagnosticResponse, error) {
	// Validar entrada
	if aprobacion != "Si" && aprobacion != "No" {
		return &models.DiagnosticResponse{
			Success: false,
			Message: "La aprobación debe ser 'Si' o 'No'",
		}, nil
	}

	if comentario == "" {
		return &models.DiagnosticResponse{
			Success: false,
			Message: "El comentario es requerido",
		}, nil
	}

	// Enviar solicitud al servicio de prediagnóstico
	result, err := s.client.CreateDiagnostic(prediagnosticID, aprobacion, comentario)
	if err != nil {
		return &models.DiagnosticResponse{
			Success: false,
			Message: fmt.Sprintf("Error al crear diagnóstico: %v", err),
		}, nil
	}

	// Procesar respuesta
	success, ok := result["success"].(bool)
	if !ok {
		// Si no hay campo "success", inferir el éxito basado en el mensaje
		success = false
	}

	message, ok := result["message"].(string)
	if !ok {
		message = "Diagnóstico procesado"
	}

	// Si el mensaje indica éxito pero success es false, corregir
	if !success && (message == "Diagnostic saved successfully" ||
		message == "Diagnóstico guardado exitosamente" ||
		message == "Diagnostic created successfully") {
		success = true
	}

	diagnosticID, _ := result["diagnostic_id"].(string)

	return &models.DiagnosticResponse{
		Success:      success,
		Message:      message,
		DiagnosticID: diagnosticID,
	}, nil
}

// CreateDiagnosticWithAutoNotification obtiene datos del caso automáticamente y envía notificación
func (s *DiagnosticService) CreateDiagnosticWithAutoNotification(prediagnosticID, aprobacion, comentario string) (*models.DiagnosticResponse, error) {
	// Crear el diagnóstico primero
	response, err := s.CreateDiagnostic(prediagnosticID, aprobacion, comentario)
	if err != nil {
		return response, err
	}

	// Solo enviar notificación si el diagnóstico fue creado exitosamente
	if response.Success {
		// Obtener información del caso para la notificación
		caseInfo, err := s.client.GetPreDiagnostic(prediagnosticID)
		if err != nil {
			fmt.Printf("Advertencia: No se pudo obtener info del caso para notificación: %v\n", err)
			return response, nil // No fallar el diagnóstico por fallo en notificación
		}

		// Extraer datos del caso
		userID := s.extractStringFromCase(caseInfo, "user_id", "")

		// Obtener información real del usuario desde el AuthService
		userInfo, err := s.authClient.GetUserInfo(userID)
		if err != nil {
			fmt.Printf("Advertencia: No se pudo obtener info del usuario %s: %v\n", userID, err)
			// Usar valores por defecto si falla la consulta de usuario
			userInfo = &clients.UserInfo{
				NombreCompleto: "Usuario",
				Correo:         "usuario@example.com",
			}
		} else {
			// Imprimir userInfo como JSON para debugging
			fmt.Printf("🔍 UserInfo obtenido: %+v\n", userInfo)
			fmt.Printf("📧 Email: %s, 👤 Nombre: %s, 🆔 ID: %s\n",
				userInfo.Correo, userInfo.NombreCompleto, userInfo.Identificacion)
		}

		// Enviar notificación de forma asíncrona
		go func() {
			if err := s.notificationClient.SendDiagnosticReadyNotification(
				userInfo.Identificacion,
				userInfo.Correo,
				userInfo.NombreCompleto,
			); err != nil {
				fmt.Printf("Error enviando notificación para diagnóstico %s: %v\n", prediagnosticID, err)
			} else {
				fmt.Printf("Notificación enviada exitosamente para diagnóstico %s\n", prediagnosticID)
			}
		}()
	}

	return response, nil
}

// Función auxiliar para extraer strings de manera segura
func (s *DiagnosticService) extractStringFromCase(caseData map[string]interface{}, field, defaultValue string) string {
	if value, exists := caseData[field]; exists && value != nil {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}
