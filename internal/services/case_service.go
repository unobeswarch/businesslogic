package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/unobeswarch/businesslogic/internal/clients"
	"github.com/unobeswarch/businesslogic/internal/graph/model"
)

type CaseService struct {
	prediagnosticClient *clients.PreDiagnosticClient
}

// GetCasesByUserID obtiene los casos del usuario desde el servicio prediagnostic
func (s *CaseService) GetCasesByUserID(userID string) ([]*model.Case, error) {
	rawCases, err := s.prediagnosticClient.GetCasesByUserID(userID)
	if err != nil {
		if err.Error() == "no radiografias" {
			return nil, fmt.Errorf("no radiografias")
		}
		return nil, err
	}
	var cases []*model.Case
	for _, rawCase := range rawCases {
		processedCase, err := s.processAndStandardizeCase(rawCase)
		if err != nil {
			continue
		}
		cases = append(cases, processedCase)
	}
	return cases, nil
}

func NewCaseService(prediagnosticURL string) *CaseService {
	return &CaseService{
		prediagnosticClient: clients.NewPrediagnosticClient(prediagnosticURL),
	}
}

// GetAllCases obtiene todos los casos y los procesa/estandariza
func (s *CaseService) GetAllCases() ([]*model.Case, error) {
	// Obtener datos raw del servicio prediagnostic
	rawCases, err := s.prediagnosticClient.GetCases()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo casos del servicio prediagnostic: %w", err)
	}

	// Procesar y estandarizar los datos
	var cases []*model.Case
	for _, rawCase := range rawCases {
		processedCase, err := s.processAndStandardizeCase(rawCase)
		if err != nil {
			// Log error pero continúa procesando otros casos
			fmt.Printf("Error procesando caso: %v\n", err)
			continue
		}
		cases = append(cases, processedCase)
	}

	return cases, nil
}

// processAndStandardizeCase transforma los datos raw en formato legible y estandarizado
func (s *CaseService) processAndStandardizeCase(rawCase map[string]interface{}) (*model.Case, error) {
	// Extraer ID del caso (viene como prediagnostico_id)
	caseID, ok := rawCase["prediagnostico_id"].(string)
	if !ok {
		if id, exists := rawCase["id"]; exists {
			if idStr, isStr := id.(string); isStr {
				caseID = idStr
			} else {
				return nil, fmt.Errorf("ID del caso no válido")
			}
		} else {
			return nil, fmt.Errorf("ID del caso no encontrado")
		}
	}

	// Para user_id, usar el userID del JWT context (se obtiene del resolver)
	// No viene en los datos del prediagnostic service
	pacienteID := "3" // Default value, will be overridden by resolver context

	// Extraer información del paciente
	pacienteNombre := s.extractStringField(rawCase, "paciente_nombre", "Test Patient GUI")
	pacienteEmail := "patient.gui@test.com" // Default value

	// Extraer fecha de subida (viene como "fecha")
	fechaSubida := s.processDate(rawCase["fecha"])

	// Extraer y procesar estado
	estado := s.processStatus(rawCase["estado"])

	// URL de radiografía - construir URL real desde radiografia_ruta
	imagenPath := s.extractStringField(rawCase, "radiografia_ruta", "")
	var urlRadiografia string
	if imagenPath != "" {
		// Extract filename from full path (e.g., "storage\\radiografias\\RAD-xxx.jpg" -> "RAD-xxx.jpg")
		pathParts := strings.Split(imagenPath, "\\")
		if len(pathParts) > 0 {
			filename := pathParts[len(pathParts)-1]
			urlRadiografia = fmt.Sprintf("http://localhost:8000/prediagnostic/image/%s", filename)
		} else {
			urlRadiografia = "/placeholder-radiography.jpg"
		}
	} else {
		urlRadiografia = "/placeholder-radiography.jpg"
	}

	// Extraer doctor asignado (puede ser nil)
	doctorAsignado := s.extractStringField(rawCase, "doctor_asignado", "")

	// Procesar resultados del modelo - datos vienen directamente en el response
	var resultados *model.ResultadosModelo
	if _, exists := rawCase["probabilidad"]; exists {
		resultados = &model.ResultadosModelo{
			ProbNeumonia:       s.extractFloatField(rawCase, "probabilidad"),
			Etiqueta:           s.extractStringField(rawCase, "diagnostico_ia", "Sin diagnóstico"),
			FechaProcesamiento: fechaSubida, // Use upload date as processing date
		}
	}

	return &model.Case{
		ID:             caseID,
		PacienteID:     pacienteID,
		PacienteNombre: pacienteNombre,
		PacienteEmail:  pacienteEmail,
		FechaSubida:    fechaSubida,
		Estado:         estado,
		URLRadiografia: urlRadiografia,
		Resultados:     resultados,
		DoctorAsignado: &doctorAsignado,
	}, nil
}

// Funciones auxiliares para extraer y procesar campos

func (s *CaseService) extractStringField(data map[string]interface{}, field string, defaultValue string) string {
	if value, exists := data[field]; exists && value != nil {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

func (s *CaseService) extractFloatField(data map[string]interface{}, field string) float64 {
	if value, exists := data[field]; exists && value != nil {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
	}
	return 0.0
}

func (s *CaseService) processDate(dateValue interface{}) string {
	if dateValue == nil {
		return "Fecha no disponible"
	}

	if dateStr, ok := dateValue.(string); ok {
		// Intentar parsear y reformatear la fecha para hacerla más legible
		if parsedTime, err := time.Parse(time.RFC3339, dateStr); err == nil {
			return parsedTime.Format("02/01/2006 15:04")
		}
		// Si no se puede parsear, devolver tal como está
		return dateStr
	}

	return "Fecha no disponible"
}

func (s *CaseService) processStatus(statusValue interface{}) string {
	if statusValue == nil {
		return "Estado desconocido"
	}

	if statusStr, ok := statusValue.(string); ok {
		// Estandarizar estados a formato legible
		switch statusStr {
		case "pending":
			return "Pendiente"
		case "processing":
			return "En procesamiento"
		case "completed":
			return "Completado"
		case "error":
			return "Error"
		case "reviewed":
			return "Revisado"
		default:
			return statusStr
		}
	}

	return "Estado desconocido"
}

func (s *CaseService) processLabel(labelValue interface{}) string {
	if labelValue == nil {
		return "Sin clasificar"
	}

	if labelStr, ok := labelValue.(string); ok {
		// Estandarizar etiquetas a formato legible
		switch labelStr {
		case "pneumonia":
			return "Neumonía detectada"
		case "normal":
			return "Normal"
		case "uncertain":
			return "Resultado incierto"
		default:
			return labelStr
		}
	}

	return "Sin clasificar"
}

// GetCaseDetail obtiene información COMPLETA de UNA radiografía específica (HU7)
// Este método es llamado por el GraphQL resolver CaseDetail
//
// Flujo completo para HU7:
// 1. GraphQL resolver → CaseService.GetCaseDetail(caseID, userID)
// 2. Validar que el caso pertenece al usuario (security)
// 3. REST call → prediagnostic/case/{caseID} para datos básicos
// 4. Si estado="validado" → REST call prediagnostic/diagnostic/{caseID}
// 5. Consolidar datos → GraphQL CaseDetail model
//
// Parámetros:
//   - caseID: ID del caso/radiografía a obtener detalles
//   - userID: ID del usuario autenticado (para validar propiedad)
//
// Retorna: *model.CaseDetail con información completa o error
func (s *CaseService) GetCaseDetail(caseID, userID string) (*model.CaseDetail, error) {
	// PASO 1: Obtener información básica del caso
	caseData, err := s.prediagnosticClient.GetPreDiagnostic(caseID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo caso del servicio prediagnostic: %w", err)
	}

	// PASO 2: Validar propiedad del caso (SEGURIDAD CRÍTICA)
	// El paciente solo puede ver SUS propios casos
	if !s.validateCaseOwnership(caseData, userID) {
		return nil, fmt.Errorf("acceso denegado: caso no pertenece al usuario")
	}

	// PASO 3: Procesar datos básicos directamente
	// Extraer ID del caso
	prediagnosticoID := s.extractStringField(caseData, "prediagnostico_id", "")
	if prediagnosticoID == "" {
		return nil, fmt.Errorf("ID del caso no válido")
	}

	// Extraer user_id
	userIDFromData := s.extractStringField(caseData, "user_id", "")
	if userIDFromData == "" {
		return nil, fmt.Errorf("ID del usuario no válido")
	}

	// Procesar estado
	estado := s.extractStringField(caseData, "estado", "Pendiente")

	// Procesar fechas
	fechaSubida := s.processDate(caseData["fecha_subida"])

	// Procesar resultados del modelo
	var resultados *model.ResultadosModelo
	if resultadosRaw, exists := caseData["resultado_modelo"]; exists && resultadosRaw != nil {
		if resultadosMap, ok := resultadosRaw.(map[string]interface{}); ok {
			resultados = &model.ResultadosModelo{
				ProbNeumonia:       s.extractFloatField(resultadosMap, "probabilidad_neumonia"),
				Etiqueta:           s.extractStringField(resultadosMap, "etiqueta", "No disponible"),
				FechaProcesamiento: s.processDate(caseData["fecha_procesamiento"]),
			}
		}
	}

	// Construir URL de radiografía
	radiografiaRuta := s.extractStringField(caseData, "radiografia_ruta", "")
	urlRadiografia := ""
	if radiografiaRuta != "" {
		// Extract filename from full path (e.g., "storage\\radiografias\\RAD-xxx.jpg" -> "RAD-xxx.jpg")
		pathParts := strings.Split(radiografiaRuta, "\\")
		if len(pathParts) > 0 {
			filename := pathParts[len(pathParts)-1]
			urlRadiografia = fmt.Sprintf("http://localhost:8000/prediagnostic/image/%s", filename)
		}
	}

	// PASO 4: Construir CaseDetail base
	// Crear PreDiagnostic con ResultadosModelo anidado
	preDiagnostic := &model.PreDiagnostic{
		PrediagnosticID:  prediagnosticoID,
		PacienteID:       userIDFromData,
		Urlrad:           urlRadiografia,
		Estado:           estado,
		ResultadosModelo: resultados,
		FechaSubida:      fechaSubida,
	}

	caseDetail := &model.CaseDetail{
		ID:            prediagnosticoID,
		RadiografiaID: prediagnosticoID, // mismo ID para simplificar
		URLImagen:     urlRadiografia,
		Estado:        estado,
		FechaSubida:   fechaSubida,
		PreDiagnostic: preDiagnostic, // PreDiagnostic completo
		Diagnostic:    nil,           // Se llena si existe
	}

	// PASO 5: Obtener diagnóstico médico si el caso está validado
	if estado == "Validado" {
		diagnostic, err := s.getDiagnosticForCase(caseID)
		if err != nil {
			// Log warning pero continuar - diagnóstico es opcional
			log.Printf("Warning: no se pudo obtener diagnóstico para caso %s: %v", caseID, err)
		} else {
			caseDetail.Diagnostic = diagnostic
		}
	}

	return caseDetail, nil
}

// validateCaseOwnership valida que el caso pertenece al usuario autenticado
// Función de SEGURIDAD - evita que pacientes vean casos de otros
func (s *CaseService) validateCaseOwnership(caseData map[string]interface{}, userID string) bool {
	// Extraer user_id del caso de manera segura
	caseUserID, ok := caseData["user_id"].(string)
	if !ok {
		log.Printf("Error: caso sin user_id válido")
		return false
	}

	return caseUserID == userID
}

// getDiagnosticForCase obtiene diagnóstico médico si existe
// Llamada REST interna al servicio Python
func (s *CaseService) getDiagnosticForCase(caseID string) (*model.Diagnostic, error) {
	// REST call interno: GET prediagnostic/diagnostic/{caseID}
	diagnosticData, err := s.prediagnosticClient.GetDiagnostic(caseID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo diagnóstico: %w", err)
	}

	// Transformar datos raw → model GraphQL (mapping DB fields to GraphQL schema)
	doctorNombre := getString(diagnosticData, "doctor_nombre")
	diagnostic := &model.Diagnostic{
		ID:               getDiagnosticID(diagnosticData),
		PrediagnosticoID: getString(diagnosticData, "case_id"),          // DB field "case_id" → GraphQL "prediagnosticoId"
		Aprobacion:       getString(diagnosticData, "validacion"),       // DB field "validacion" → GraphQL "aprobacion"
		Comentarios:      getString(diagnosticData, "diagnostico"),      // DB field "diagnostico" → GraphQL "comentarios"
		FechaRevision:    getString(diagnosticData, "fecha_validacion"), // DB field "fecha_validacion" → GraphQL "fechaRevision"
		DoctorNombre:     &doctorNombre,                                 // DB field "doctor_nombre" → GraphQL "doctorNombre"
	}

	return diagnostic, nil
}

// getString extrae string de map[string]interface{} de manera segura
// Función helper para convertir datos JSON → GraphQL models
func getString(data map[string]interface{}, field string) string {
	if value, exists := data[field]; exists && value != nil {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

// getDiagnosticID extrae ID del diagnóstico de manera segura
// Maneja tanto string como ObjectId formats
func getDiagnosticID(data map[string]interface{}) string {
	if id, exists := data["id"]; exists && id != nil {
		if strValue, ok := id.(string); ok {
			return strValue
		}
	}
	// Fallback a _id si no hay id
	if id, exists := data["_id"]; exists && id != nil {
		if strValue, ok := id.(string); ok {
			return strValue
		}
	}
	return ""
}
