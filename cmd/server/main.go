package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/unobeswarch/businesslogic/internal/graph"
	"github.com/unobeswarch/businesslogic/internal/graph/generated"
	"github.com/unobeswarch/businesslogic/internal/handlers"
	"github.com/unobeswarch/businesslogic/internal/services"
)

const defaultPort = "8080"

func main() {
	port := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	// URL del servicio de prediagnóstico (configurable por variable de entorno)
	prediagnosticURL := os.Getenv("PREDIAGNOSTIC_SERVICE_URL")
	if prediagnosticURL == "" {
		prediagnosticURL = "http://localhost:8000" // URL por defecto
	}

	// Instanciamos los services
	prediagnosticService := services.NewPrediagnosticService(prediagnosticURL)
	caseService := services.NewCaseService(prediagnosticURL)
	authService := services.NewAuthService()
	diagnosticService := services.NewDiagnosticService(prediagnosticURL)

	// Inyectamos los services en el resolver
	resolver := &graph.Resolver{
		PrediagnosticSrv: prediagnosticService,
		CaseSrv:          caseService,
		AuthSrv:          authService,
		DiagnosticSrv:    diagnosticService,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Middleware para extraer Authorization header y agregarlo al contexto
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			// Extraer Authorization header
			authHeader := r.Header.Get("Authorization")

			// Agregar al contexto si existe
			ctx := r.Context()
			if authHeader != "" {
				ctx = context.WithValue(ctx, "Authorization", authHeader)
			}

			// Continuar con el siguiente handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", authMiddleware(srv))
	http.Handle("/register", authMiddleware(http.HandlerFunc(handlers.HandlerRegistrarUsuario)))
	http.Handle("/auth", authMiddleware(http.HandlerFunc(handlers.HandlerIniciarSesion)))
	http.Handle("/validation", authMiddleware(http.HandlerFunc(handlers.HandlerValidacion)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Printf("prediagnostic service URL: %s", prediagnosticURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
