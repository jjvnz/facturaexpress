package main

import (
	"facturaexpress/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	jwtKey := []byte(os.Getenv("SECRET_KEY"))
	expTimeStr := os.Getenv("EXP_TIME")

	// Crea un nuevo enrutador Gin y configura las rutas y los controladores de ruta
	router := routes.NewRouter(jwtKey, expTimeStr)

	// Inicia el servidor Gin y escucha las solicitudes entrantes
	router.Run("0.0.0.0:8000")
}
