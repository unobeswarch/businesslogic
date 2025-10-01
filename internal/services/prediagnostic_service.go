package services

import (
	"github.com/unobeswarch/businesslogic/internal/clients"
	"github.com/unobeswarch/businesslogic/internal/graph/model"
)

type PreDiagnosticService struct {
	client *clients.PreDiagnosticClient
}

func NewPrediagnosticService(baseURL string) *PreDiagnosticService {
	return &PreDiagnosticService{
		client: clients.NewPrediagnosticClient(baseURL),
	}
}

// GetPreDiagnosticByID busca un prediagnóstico por ID
func (s *PreDiagnosticService) GetPreDiagnosticByID(id string) (*model.PreDiagnostic, error) {
	data, err := s.client.GetPreDiagnostic(id)
	if err != nil {
		return nil, err
	}

	// Mapear JSON -> GraphQL model (usando los nombres correctos del JSON)
	resultados := data["resultado_modelo"].(map[string]interface{})
	return &model.PreDiagnostic{
		PrediagnosticID: id,
		PacienteID:      data["user_id"].(string),
		Urlrad:          data["radiografia_ruta"].(string),
		Estado:          data["estado"].(string),
		ResultadosModelo: &model.ResultadosModelo{
			ProbNeumonia:       resultados["probabilidad_neumonia"].(float64),
			Etiqueta:           resultados["etiqueta"].(string),
			FechaProcesamiento: data["fecha_procesamiento"].(string),
		},
		FechaSubida: data["fecha_subida"].(string),
	}, nil
}
