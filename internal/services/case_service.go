package services

import (
	"fmt"
	"time"

	"github.com/unobeswarch/businesslogic/internal/clients"
	"github.com/unobeswarch/businesslogic/internal/graph/model"
)

type CaseService struct {
	prediagnosticClient *clients.PreDiagnosticClient
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
	// Extraer ID del caso
	caseID, ok := rawCase["id"].(string)
	if !ok {
		if idFloat, isFloat := rawCase["id"].(float64); isFloat {
			caseID = fmt.Sprintf("%.0f", idFloat)
		} else {
			return nil, fmt.Errorf("ID del caso no válido")
		}
	}

	// Extraer ID del paciente
	pacienteID, ok := rawCase["user_id"].(string)
	if !ok {
		if idFloat, isFloat := rawCase["user_id"].(float64); isFloat {
			pacienteID = fmt.Sprintf("%.0f", idFloat)
		} else {
			return nil, fmt.Errorf("ID del paciente no válido")
		}
	}

	// Extraer información del paciente
	pacienteNombre := s.extractStringField(rawCase, "paciente_nombre", "Nombre no disponible")
	pacienteEmail := s.extractStringField(rawCase, "paciente_email", "Email no disponible")

	// Extraer fecha de subida y formatearla
	fechaSubida := s.processDate(rawCase["fecha_subida"])

	// Extraer y procesar estado
	estado := s.processStatus(rawCase["estado"])

	// Extraer URL de radiografía
	urlRadiografia := s.extractStringField(rawCase, "radiografia_url", "")

	// Extraer doctor asignado (puede ser nil)
	doctorAsignado := s.extractStringField(rawCase, "doctor_asignado", "")

	// Procesar resultados del modelo si existen
	var resultados *model.ResultadosModelo
	if resultadosRaw, exists := rawCase["resultado_modelo"]; exists && resultadosRaw != nil {
		if resultadosMap, ok := resultadosRaw.(map[string]interface{}); ok {
			resultados = &model.ResultadosModelo{
				ProbNeumonia:       s.extractFloatField(resultadosMap, "probabilidad_neumonia"),
				Etiqueta:          s.processLabel(resultadosMap["etiqueta"]),
				FechaProcesamiento: s.processDate(rawCase["fecha_procesamiento"]),
			}
		}
	}

	return &model.Case{
		ID:              caseID,
		PacienteID:      pacienteID,
		PacienteNombre:  pacienteNombre,
		PacienteEmail:   pacienteEmail,
		FechaSubida:     fechaSubida,
		Estado:          estado,
		URLRadiografia:  urlRadiografia,
		Resultados:      resultados,
		DoctorAsignado:  &doctorAsignado,
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