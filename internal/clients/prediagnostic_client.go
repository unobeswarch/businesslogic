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

// GetCasesByUserID obtiene los casos del usuario desde el servicio prediagnostic
func (c *PreDiagnosticClient) GetCasesByUserID(userID string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/prediagnostic/cases/%s", c.BaseURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no radiografias")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("respuesta vacía del servidor")
	}

	// El servicio Python devuelve {cases: [...]} no directamente [...]
	var responseWrapper struct {
		Cases []map[string]interface{} `json:"cases"`
	}
	if err := json.Unmarshal(body, &responseWrapper); err != nil {
		return nil, fmt.Errorf("error parseando JSON de casos: %w", err)
	}

	return responseWrapper.Cases, nil
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

	// Convertir "Si"/"No" a boolean para el servicio externo
	var aprobacionBool bool
	if aprobacion == "Si" {
		aprobacionBool = true
	} else {
		aprobacionBool = false
	}

	// Preparar el payload
	payload := map[string]interface{}{
		"prediagnostic_id": prediagnosticID,
		"aprobacion":       aprobacionBool, // Enviamos boolean al servicio externo
		"comentario":       comentario,
		"fecha_revision":   fmt.Sprintf("%d", time.Now().Unix()), // timestamp actual
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		return nil, err
	}

	// Debug: Mostrar URL y payload
	fmt.Printf("Enviando POST a: %s\n", url)
	fmt.Printf("Payload enviado: %s\n", string(jsonPayload))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error en la petición POST: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Leer el body antes de verificar el código de estado para obtener más información del error
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error leyendo el body de diagnóstico: %v\n", err)
		return nil, err
	}

	// Debug: Mostrar el contenido de la respuesta
	fmt.Printf("Respuesta de diagnóstico del servidor (Status %d): %s\n", resp.StatusCode, string(body))

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Intentar parsear el error del body
		var errorResult map[string]interface{}
		if err := json.Unmarshal(body, &errorResult); err == nil {
			if detail, ok := errorResult["detail"]; ok {
				return nil, fmt.Errorf("respuesta HTTP %d: %v", resp.StatusCode, detail)
			}
		}
		return nil, fmt.Errorf("respuesta HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parseando JSON de diagnóstico: %v\n", err)
		return nil, fmt.Errorf("error parseando JSON de diagnóstico: %w", err)
	}

	return result, nil
}

// GetCase obtiene información básica de UN caso específico (HU7)
// Llamada REST: GET /case/{caseID}
//
// Parámetros:
//   - caseID: ID del caso a obtener
//
// Retorna: map[string]interface{} con datos del caso o error
func (c *PreDiagnosticClient) GetCase(caseID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/case/%s", c.BaseURL, caseID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error en petición HTTP para caso %s: %w", caseID, err)
	}
	defer resp.Body.Close()

	// Verificar el código de estado
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta HTTP %d para caso %s: %s", resp.StatusCode, caseID, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta para caso %s: %w", caseID, err)
	}

	// Verificar si el body está vacío
	if len(body) == 0 {
		return nil, fmt.Errorf("respuesta vacía del servidor para caso %s", caseID)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parseando JSON para caso %s: %w", caseID, err)
	}

	return result, nil
}

// GetDiagnostic obtiene diagnóstico médico de un caso (HU7)
// Llamada REST: GET /diagnostic/{caseID}
//
// Parámetros:
//   - caseID: ID del caso para obtener diagnóstico
//
// Retorna: map[string]interface{} con diagnóstico médico o error
func (c *PreDiagnosticClient) GetDiagnostic(caseID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/diagnostic/%s", c.BaseURL, caseID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error en petición HTTP para diagnóstico %s: %w", caseID, err)
	}
	defer resp.Body.Close()

	// 404 es válido - no todos los casos tienen diagnóstico
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("diagnóstico no encontrado para caso %s", caseID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("respuesta HTTP %d para diagnóstico %s: %s", resp.StatusCode, caseID, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta para diagnóstico %s: %w", caseID, err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("respuesta vacía del servidor para diagnóstico %s", caseID)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parseando JSON para diagnóstico %s: %w", caseID, err)
	}

	return result, nil
}
