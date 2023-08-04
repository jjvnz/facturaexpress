package main

import (
	"facturaexpress/data"
	"facturaexpress/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Carga los datos del archivo .env
	loadDataEnv()

	// Carga la clave secreta para firmar los tokens JWT desde la variable de entorno
	jwtKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	// Crea una nueva instancia de *data.DB y conéctate a la base de datos
	db, err := data.NewDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Crea un nuevo enrutador Gin y configura las rutas y los controladores de ruta
	router := routes.NewRouter(db, jwtKey)

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
