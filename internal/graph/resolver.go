package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.
import (
	"github.com/unobeswarch/businesslogic/internal/services"
)

type Resolver struct {
	PrediagnosticSrv *services.PreDiagnosticService
	CaseSrv          *services.CaseService
	AuthSrv          *services.AuthService
	DiagnosticSrv    *services.DiagnosticService
}
