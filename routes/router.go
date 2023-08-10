package routes

import (
	"facturaexpress/common"
	"facturaexpress/data"
	handler "facturaexpress/handlers"
	middleware "facturaexpress/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(db *data.DB, jwtKey []byte, expTimeStr string) *gin.Engine {
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"192.168.1.2", "10.0.0.0/8"})

	// Configurar la política de CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	router.Use(cors.New(config))

	v1 := router.Group("/v1")
	{
		v1.POST("/register", func(c *gin.Context) {
			handler.Register(c, db)
		})

		v1.POST("/login", func(c *gin.Context) {
			handler.Login(c, db, jwtKey, expTimeStr)
		})

		// Rutas protegidas con el middleware AuthMiddleware
		authorized := v1.Group("/")
		authorized.Use(func(c *gin.Context) {
			middleware.AuthMiddleware(c, db, jwtKey)
		})
		{
			adminRoutes := authorized.Group("/")
			adminRoutes.Use(func(c *gin.Context) {
				middleware.RoleAuthMiddleware(c, db, common.ADMIN)
			})

			adminRoutes.PUT("/users/:userID/roles/:roleID", func(c *gin.Context) {
				handler.AssignRole(c, db)
			})

			adminRoutes.PUT("/users/:userID/new-role/:newRoleID", func(c *gin.Context) {
				handler.ActualizarRol(c, db)
			})

			adminRoutes.GET("/roles", func(c *gin.Context) {
				handler.ListRoles(c, db)
			})

			adminRoutes.GET("/usuarios", func(c *gin.Context) {
				handler.ListarUsuarios(c, db)
			})
			adminRoutes.POST("/usuarios", func(c *gin.Context) {
				handler.CrearUsuario(c, db)
			})
			adminRoutes.PUT("/usuarios/:id", func(c *gin.Context) {
				handler.ActualizarUsuario(c, db)
			})
			adminRoutes.DELETE("/usuarios/:id", func(c *gin.Context) {
				handler.EliminarUsuario(c, db)
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

			// En tu archivo router.go, agrega una nueva ruta para manejar solicitudes de logout
			authorized.POST("/logout", func(c *gin.Context) {
				handler.Logout(c, db)
			})

		}
	}

	return router
}
