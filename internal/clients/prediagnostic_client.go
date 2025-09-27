package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PreDiagnosticClient struct {
	BaseURL string
}

func NewPrediagnosticClient(baseURL string) *PreDiagnosticClient {
	return &PreDiagnosticClient{BaseURL: baseURL}
}

func (c *PreDiagnosticClient) GetPreDiagnostic(id string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/prediagnostic/case/%s", c.BaseURL, id)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error en la petición HTTP: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error leyendo el body: %v\n", err)
		return nil, err
	}

	// Debug: Mostrar el contenido de la respuesta
	fmt.Printf("Respuesta del servidor: %s\n", string(body))

	// Verificar si el body está vacío
	if len(body) == 0 {
		return nil, fmt.Errorf("respuesta vacía del servidor")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parseando JSON: %v\n", err)
		return nil, fmt.Errorf("error parseando JSON: %w", err)
	}

	// Verificar si el resultado es nil
	if result == nil {
		return nil, fmt.Errorf("el servidor devolvió null")
	}

	return result, nil
}

// GetCases obtiene todos los casos del servicio de prediagnóstico
func (c *PreDiagnosticClient) GetCases() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/prediagnostic/cases", c.BaseURL)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error en la petición HTTP para obtener casos: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error leyendo el body de casos: %v\n", err)
		return nil, err
	}

	// Debug: Mostrar el contenido de la respuesta
	fmt.Printf("Respuesta de casos del servidor: %s\n", string(body))

	// Verificar si el body está vacío
	if len(body) == 0 {
		return nil, fmt.Errorf("respuesta vacía del servidor")
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parseando JSON de casos: %v\n", err)
		return nil, fmt.Errorf("error parseando JSON de casos: %w", err)
	}

	return result, nil
}