package main

import (
	"facturaexpress/api"
	"facturaexpress/pkg/storage"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Carga los datos del archivo .env
	loadDataEnv()

	// Crea una nueva instancia de *storage.DB y conéctate a la base de datos
	db, err := storage.NewDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Crea un nuevo enrutador Gin y configura las rutas y los controladores de ruta
	router := api.NewRouter(db)

	// Inicia el servidor Gin y escucha las solicitudes entrantes
	router.Run(":8000")
}

// Carga los datos del archivo .env
func loadDataEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error al cargar el archivo .env: %v", err)
	}
}
