package router

import (
	"os"
	"taskmanager/controllers"
	"taskmanager/data"
	"taskmanager/middleware"
	"taskmanager/models"

	"github.com/gin-gonic/gin"
)

func SetupRouter(ts *data.TaskService, us *data.UserService) *gin.Engine {
	// itialize task and user controller
	taskController := controllers.NewTaskController(ts)
	userController := controllers.NewUserController(us)

	// intialize the router
	router := gin.Default()

	// group routes under /api/v1
	api := router.Group("/api/v1")

	// task routes
	taskRoutes := api.Group("/tasks")

	// user routes
	userRoutes := api.Group("/user")

	// read jwt secret from env varaiable
	jwtSecret := os.Getenv("JWT_SECRET")

	// routes only need authentication
	taskRoutes.GET("", middleware.AuthMiddleware(jwtSecret), taskController.GetTasks)
	taskRoutes.GET("/:id", middleware.AuthMiddleware(jwtSecret), taskController.GetTaskById)

	// admin-only routes
	adminTaskRoutes := taskRoutes.Group("/")

	adminTaskRoutes.Use(middleware.AuthMiddleware(jwtSecret))
	adminTaskRoutes.Use(middleware.AuthorizationMiddleware(models.RoleAdmin))

	adminTaskRoutes.POST("", taskController.CreatTask)
	adminTaskRoutes.PUT("/:id", taskController.UpdateTask)
	adminTaskRoutes.DELETE("/:id", taskController.DeleteTask)

	userRoutes.POST("/register", userController.RegisterUser)
	userRoutes.POST("/login", userController.AuthenticateUser)
	userRoutes.PATCH("/:id/promote", middleware.AuthMiddleware(jwtSecret), middleware.AuthorizationMiddleware(models.RoleAdmin), userController.PromoteUser)

	return router
}
