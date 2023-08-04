package routes

import (
	"facturaexpress/data"
	handler "facturaexpress/handlers"
	middleware "facturaexpress/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(db *data.DB, jwtKey []byte) *gin.Engine {
	router := gin.Default()

	// Configurar la política de CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	router.Use(cors.New(config))

	router.POST("/register", func(c *gin.Context) {
		handler.Register(c, db)
	})

	router.POST("/login", func(c *gin.Context) {
		handler.Login(c, db, jwtKey)
	})

	// Rutas protegidas con el middleware AuthMiddleware
	authorized := router.Group("/")
	authorized.Use(func(c *gin.Context) {
		middleware.AuthMiddleware(c, db, jwtKey)
	})
	{
		adminRoutes := authorized.Group("/")
		adminRoutes.Use(func(c *gin.Context) {
			middleware.RoleAuthMiddleware(c, db, "administrador")
		})

		adminRoutes.PUT("/users/:userID/roles/:roleID", func(c *gin.Context) {
			handler.AssignRole(c, db)
		})

		adminRoutes.PUT("/users/:userID/new_role/:newRoleID", func(c *gin.Context) {
			handler.ActualizarRol(c, db)
		})

		adminRoutes.GET("/roles", func(c *gin.Context) {
			handler.ListRoles(c, db)
		})

		authorized.GET("/facturas", func(c *gin.Context) {
			handler.ListarFacturas(c, db)
		})

		authorized.POST("/facturas", func(c *gin.Context) {
			handler.CrearFactura(c, db)
		})

		authorized.PUT("/facturas/:id", func(c *gin.Context) {
			handler.ActualizarFactura(c, db)
		})

		authorized.DELETE("/facturas/:id", func(c *gin.Context) {
			handler.EliminarFactura(c, db)
		})

		// Agregar nueva ruta para generar PDFs
		authorized.GET("/facturas/:id/pdf", func(c *gin.Context) {
			handler.GenerarPDF(c, db)
		})
	}

	return router
}
