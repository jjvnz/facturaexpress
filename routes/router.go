package routes

import (
	"facturaexpress/common"
	authHandler "facturaexpress/handlers/auth"
	invoiceHandler "facturaexpress/handlers/invoice"
	roleHandler "facturaexpress/handlers/role"
	userHandler "facturaexpress/handlers/user"
	middleware "facturaexpress/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(jwtKey []byte, expTimeStr string) *gin.Engine {
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"192.168.1.2", "10.0.0.0/8"})

	// Configure CORS policy
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowHeaders = []string{"Authorization", "Content-Type"}
	router.Use(cors.New(config))

	v1 := router.Group("/v1")
	{
		v1.POST("/register", func(c *gin.Context) {
			authHandler.Register(c)
		})

		v1.POST("/login", func(c *gin.Context) {
			authHandler.Login(c, jwtKey, expTimeStr)
		})

		// Routes protected with AuthMiddleware middleware
		authorized := v1.Group("/")
		authorized.Use(func(c *gin.Context) {
			middleware.AuthMiddleware(c, jwtKey)
		})
		{
			adminRoutes := authorized.Group("/")
			adminRoutes.Use(func(c *gin.Context) {
				middleware.RoleAuthMiddleware(c, common.ADMIN)
			})

			adminRoutes.PUT("/users/:id/new-role/:newRoleID", func(c *gin.Context) {
				roleHandler.AssignRole(c)
			})
			adminRoutes.PUT("/users/:id/roles/:roleID", func(c *gin.Context) {
				roleHandler.UpdateRole(c)
			})

			adminRoutes.GET("/roles", func(c *gin.Context) {
				roleHandler.ListRoles(c)
			})

			adminRoutes.GET("/users", func(c *gin.Context) {
				userHandler.ListUsers(c)
			})
			adminRoutes.POST("/users", func(c *gin.Context) {
				userHandler.CreateUser(c)
			})
			adminRoutes.PUT("/users/:id", func(c *gin.Context) {
				userHandler.UpdateUser(c)
			})
			adminRoutes.DELETE("/users/:id", func(c *gin.Context) {
				userHandler.DeleteUser(c)
			})

			authorized.GET("/invoices", func(c *gin.Context) {
				invoiceHandler.ListInvoices(c)
			})

			authorized.POST("/invoices", func(c *gin.Context) {
				invoiceHandler.CreateInvoice(c)
			})

			authorized.PUT("/invoices/:id", func(c *gin.Context) {
				invoiceHandler.UpdateInvoice(c)
			})

			authorized.DELETE("/invoices/:id", func(c *gin.Context) {
				invoiceHandler.DeleteInvoice(c)
			})

			// route to generate PDFs
			authorized.GET("/invoices/:id/pdf", func(c *gin.Context) {
				invoiceHandler.GeneratePDF(c)
			})

			// route to handle logout requests
			authorized.POST("/logout", func(c *gin.Context) {
				authHandler.Logout(c)
			})
		}
	}

	return router
}
