package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AuthClient struct {
	BaseURL string
}

type UserInfo struct {
	NombreCompleto string `json:"nombre_completo"`
	Identificacion string `json:"identificacion"`
	Correo         string `json:"correo"`
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{BaseURL: baseURL}
}

// GetUserInfo obtiene información del usuario desde el servicio de autenticación
func (c *AuthClient) GetUserInfo(userID string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/getUserInfo?id=%s", c.BaseURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request to auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("error decoding user info: %w", err)
	}

	return &userInfo, nil
}
