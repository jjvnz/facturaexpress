package api

import (
	"facturaexpress/api/handlers"
	"facturaexpress/api/middleware"
	"facturaexpress/pkg/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(db *storage.DB, jwtKey []byte) *gin.Engine {
	router := gin.Default()

	// Configurar la política de CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	router.Use(cors.New(config))

	router.GET("/hiwelcome", handlers.HelloHandler)

	router.POST("/register", func(c *gin.Context) {
		handlers.Register(c, db)
	})

	router.POST("/login", func(c *gin.Context) {
		handlers.Login(c, db, jwtKey)
	})

	// Rutas protegidas con el middleware AuthMiddleware
	authorized := router.Group("/")
	authorized.Use(func(c *gin.Context) {
		middleware.AuthMiddleware(c, jwtKey)
	})
	{
		authorized.GET("/facturas", func(c *gin.Context) {
			handlers.ListarFacturas(c, db)
		})
		authorized.POST("/facturas", func(c *gin.Context) {
			handlers.CrearFactura(c, db)
		})
		authorized.PUT("/facturas/:id", func(c *gin.Context) {
			handlers.ActualizarFactura(c, db)
		})
		authorized.DELETE("/facturas/:id", func(c *gin.Context) {
			handlers.EliminarFactura(c, db)
		})

		// Agregar nueva ruta para generar PDFs
		authorized.GET("/facturas/:id/pdf", func(c *gin.Context) {
			handlers.GenerarPDF(c, db)
		})
	}

	return router
}
