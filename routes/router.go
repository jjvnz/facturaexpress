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
		v1.POST("/register", func(context *gin.Context) {
			authHandler.Register(context)
		})

		v1.POST("/login", func(context *gin.Context) {
			authHandler.Login(context, jwtKey, expTimeStr)
		})

		// Routes protected with AuthMiddleware middleware
		authorized := v1.Group("/")
		authorized.Use(func(context *gin.Context) {
			middleware.AuthMiddleware(context, jwtKey)
		})
		{
			adminRoutes := authorized.Group("/")
			adminRoutes.Use(func(context *gin.Context) {
				middleware.RoleAuthMiddleware(context, common.ADMIN)
			})

			adminRoutes.PUT("/users/:id/new-role/:newRoleID", func(context *gin.Context) {
				roleHandler.AssignRole(context)
			})
			adminRoutes.PUT("/users/:id/roles/:roleID", func(context *gin.Context) {
				roleHandler.UpdateRole(context)
			})

			adminRoutes.GET("/roles", func(context *gin.Context) {
				roleHandler.ListRoles(context)
			})

			adminRoutes.GET("/users", func(context *gin.Context) {
				userHandler.ListUsers(context)
			})
			adminRoutes.POST("/users", func(context *gin.Context) {
				userHandler.CreateUser(context)
			})
			adminRoutes.PUT("/users/:id", func(context *gin.Context) {
				userHandler.UpdateUser(context)
			})
			adminRoutes.DELETE("/users/:id", func(context *gin.Context) {
				userHandler.DeleteUser(context)
			})

			authorized.GET("/user/profile", func(context *gin.Context) {
				userHandler.GetUserInfo(context)
			})

			authorized.GET("/invoices", func(context *gin.Context) {
				invoiceHandler.ListInvoices(context)
			})

			authorized.POST("/invoices", func(context *gin.Context) {
				invoiceHandler.CreateInvoice(context)
			})

			authorized.PUT("/invoices/:id", func(context *gin.Context) {
				invoiceHandler.UpdateInvoice(context)
			})

			authorized.DELETE("/invoices/:id", func(context *gin.Context) {
				invoiceHandler.DeleteInvoice(context)
			})

			// route to generate PDFs
			authorized.GET("/invoices/:id/pdf", func(context *gin.Context) {
				invoiceHandler.GeneratePDF(context)
			})

			// route to handle logout requests
			authorized.POST("/logout", func(context *gin.Context) {
				authHandler.Logout(context)
			})
		}
	}

	return router
}
