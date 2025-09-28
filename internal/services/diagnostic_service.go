package services

import (
	"fmt"

	"github.com/unobeswarch/businesslogic/internal/clients"
	"github.com/unobeswarch/businesslogic/internal/models"
)

type DiagnosticService struct {
	client *clients.PreDiagnosticClient
}

func NewDiagnosticService(baseURL string) *DiagnosticService {
	return &DiagnosticService{
		client: clients.NewPrediagnosticClient(baseURL),
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
