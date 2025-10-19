package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type NotificationClient struct {
	BaseURL string
}

type NotificationMessage struct {
	UserID       string    `json:"identificacion"`
	PatientEmail string    `json:"patient_email"`
	PatientName  string    `json:"patient_name"`
	Timestamp    time.Time `json:"timestamp"`
	MessageType  string    `json:"message_type"`
}

func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{BaseURL: baseURL}
}

func (c *NotificationClient) SendDiagnosticReadyNotification(userID, patientEmail, patientName string) error {
	msg := NotificationMessage{
		UserID:       userID,
		PatientEmail: patientEmail,
		PatientName:  patientName,
		Timestamp:    time.Now(),
		MessageType:  "diagnostic_ready",
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	url := fmt.Sprintf("%s/notifications", c.BaseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	return nil
}
