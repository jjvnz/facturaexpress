package main

import (
	"encoding/json"
	"facturaexpress/models"
	"facturaexpress/routes"
	"log"
	"os"
)

func main() {

	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error al cargar el archivo de configuración: %v", err)
	}

	var config models.JWT
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error al leer el archivo de configuración: %v", err)
	}

	jwtKey := []byte(config.JWT.SecretKey)
	expTimeStr := config.JWT.ExpTime

	// Crea un nuevo enrutador Gin y configura las rutas y los controladores de ruta
	router := routes.NewRouter(jwtKey, expTimeStr)

	// Inicia el servidor Gin y escucha las solicitudes entrantes
	router.Run("localhost:8000")
}
