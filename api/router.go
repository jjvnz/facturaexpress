package api

import (
	"facturaexpress/api/handlers"
	"facturaexpress/pkg/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(db *storage.DB) *gin.Engine {
	router := gin.Default()

	// Configurar la política de CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	router.Use(cors.New(config))

	router.GET("/hiwelcome", handlers.HelloHandler)
	router.POST("/facturas", func(c *gin.Context) {
		handlers.CrearFactura(c, db)
	})
	router.GET("/facturas", func(c *gin.Context) {
		handlers.ListarFacturas(c, db)
	})
	router.PUT("/facturas/:id", func(c *gin.Context) {
		handlers.ActualizarFactura(c, db)
	})
	router.DELETE("/facturas/:id", func(c *gin.Context) {
		handlers.EliminarFactura(c, db)
	})

	// Agregar nueva ruta para generar PDFs
	router.GET("/facturas/:id/pdf", func(c *gin.Context) {
		handlers.GenerarPDF(c, db)
	})

	return router
}
