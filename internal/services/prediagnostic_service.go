package services

import (
	"fmt"
	"io"
	"net/http"

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

func GetAllProcessedCases() ([]byte, int, error) {
	resp, err := http.Get("http://localhost:8000/prediagnostic/cases")

	if err != nil {
		return nil, 0, fmt.Errorf("error connecting to prediagnostic service: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading response from prediagnostic service: %v", err)
	}

	return body, resp.StatusCode, nil
}

func GetPrediagnosticImage(imageFilename string) (*http.Response, error) {
	url := fmt.Sprintf("http://localhost:8000/prediagnostic/image/%s", imageFilename)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error conectando con el servicio de prediagnóstico: %v", err)
	}

	return resp, nil
}
