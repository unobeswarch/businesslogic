package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

// CreateDiagnostic envía una solicitud POST para crear un diagnóstico
func (c *PreDiagnosticClient) CreateDiagnostic(prediagnosticID, aprobacion, comentario string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/prediagnostic/diagnostic/%s", c.BaseURL, prediagnosticID)

	// Preparar el payload
	payload := map[string]interface{}{
		"prediagnostic_id": prediagnosticID,
		"aprobacion":       aprobacion,
		"comentario":       comentario,
		"fecha_revision":   fmt.Sprintf("%d", time.Now().Unix()), // timestamp actual
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error en la petición POST: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("respuesta HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error leyendo el body de diagnóstico: %v\n", err)
		return nil, err
	}

	// Debug: Mostrar el contenido de la respuesta
	fmt.Printf("Respuesta de diagnóstico del servidor: %s\n", string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parseando JSON de diagnóstico: %v\n", err)
		return nil, fmt.Errorf("error parseando JSON de diagnóstico: %w", err)
	}

	return result, nil
}