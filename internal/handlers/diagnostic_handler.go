package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/unobeswarch/businesslogic/internal/services"
)

func HandlerDiagnosticDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método no permitido"})
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Falta parámetro 'prediagnostic_id'"}`))
		return
	}

	prediagID := pathParts[len(pathParts)-1]
	body, statusCode, err := services.CaseDetail(prediagID)

	fmt.Println(prediagID)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)

}
