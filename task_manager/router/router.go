package router

import (
	"taskmanager/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// group routes under /api/v1
	api := router.Group("/api/v1")

	// task routes
	taskRoutes := api.Group("/tasks")

	taskRoutes.GET("", controllers.GetTasks)
	taskRoutes.POST("", controllers.CreatTask)
	taskRoutes.GET("/:id", controllers.GetTaskById)
	taskRoutes.PUT("/:id", controllers.UpdateTask)
	taskRoutes.DELETE("/:id", controllers.DeleteTask)

	return router
}
