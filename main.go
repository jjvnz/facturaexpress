package main

import (
	"encoding/json"
	"facturaexpress/data"
	"facturaexpress/models"
	"facturaexpress/routes"
	"log"
	"os"
)

func main() {

	// Carga los valores del archivo de configuración
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
	// Accede al campo ExpTime para obtener el valor del tiempo de expiración del token
	expTimeStr := config.JWT.ExpTime

	// Crea una nueva instancia de *data.DB y conéctate a la base de datos
	db, err := data.NewDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Crea un nuevo enrutador Gin y configura las rutas y los controladores de ruta
	router := routes.NewRouter(db, jwtKey, expTimeStr)

	// Inicia el servidor Gin y escucha las solicitudes entrantes
	router.Run("localhost:8000")
}
