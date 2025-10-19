package main

import (
	"fmt"
	"log"

	"github.com/unobeswarch/businesslogic/internal/clients"
)

func main() {
	// Crear cliente AuthService
	authClient := clients.NewAuthClient("http://localhost:8081")

	// ID de usuario de prueba - cambia este valor por un ID real que tengas en tu AuthService
	testUserID := "1" // ← CAMBIA ESTE VALOR POR UN ID REAL

	fmt.Printf("🧪 Probando conexión con AuthService...\n")
	fmt.Printf("URL: http://localhost:8081/getUserInfo?id=%s\n\n", testUserID)

	// Hacer la llamada
	userInfo, err := authClient.GetUserInfo(testUserID)
	if err != nil {
		log.Printf("❌ Error obteniendo info del usuario: %v\n", err)
		return
	}

	// Mostrar resultados
	fmt.Printf("✅ ¡Éxito! Datos obtenidos:\n")
	fmt.Printf("   📧 Correo: %s\n", userInfo.Correo)
	fmt.Printf("   👤 Nombre: %s\n", userInfo.NombreCompleto)
	fmt.Printf("   🆔 ID: %s\n", userInfo.Identificacion)
}
