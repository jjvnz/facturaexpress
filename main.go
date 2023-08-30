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

	router := routes.NewRouter(jwtKey, expTimeStr)

	router.Run("localhost:8000")
}
