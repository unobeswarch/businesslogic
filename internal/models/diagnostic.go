package models

import "time"

// Diagnostic representa un diagnóstico realizado por un doctor
type Diagnostic struct {
	ID               string    `json:"_id,omitempty" bson:"_id,omitempty"`
	PrediagnosticID  string    `json:"prediagnostic_id" bson:"prediagnostic_id"`
	Aprobacion       bool      `json:"aprobacion" bson:"aprobacion"`
	Comentario       string    `json:"comentario" bson:"comentario"`
	FechaRevision    time.Time `json:"fecha_revision" bson:"fecha_revision"`
}

// DiagnosticInput representa los datos de entrada para crear un diagnóstico
type DiagnosticInput struct {
	Aprobacion string `json:"aprobacion" validate:"required,oneof=Si No"`
	Comentario string `json:"comentario" validate:"required"`
}

// DiagnosticResponse representa la respuesta al crear un diagnóstico
type DiagnosticResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DiagnosticID string `json:"diagnostic_id,omitempty"`
}
