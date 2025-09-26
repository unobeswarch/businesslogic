package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/unobeswarch/businesslogic/internal/graph"
	"github.com/unobeswarch/businesslogic/internal/graph/generated"
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

	// Instanciamos el service con el cliente
	prediagnosticService := services.NewPrediagnosticService(prediagnosticURL)

	// Inyectamos el service en el resolver
	resolver := &graph.Resolver{
		PrediagnosticSrv: prediagnosticService,
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Printf("prediagnostic service URL: %s", prediagnosticURL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
