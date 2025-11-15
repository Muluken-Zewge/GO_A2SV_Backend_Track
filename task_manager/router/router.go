package router

import (
	"taskmanager/controllers"
	"taskmanager/data"

	"github.com/gin-gonic/gin"
)

func SetupRouter(ts *data.TaskService) *gin.Engine {
	// itialize task controller
	taskController := controllers.NewTaskController(ts)

	// intialize the router
	router := gin.Default()

	// group routes under /api/v1
	api := router.Group("/api/v1")

	// task routes
	taskRoutes := api.Group("/tasks")

	taskRoutes.GET("", taskController.GetTasks)
	taskRoutes.POST("", taskController.CreatTask)
	taskRoutes.GET("/:id", taskController.GetTaskById)
	taskRoutes.PUT("/:id", taskController.UpdateTask)
	taskRoutes.DELETE("/:id", taskController.DeleteTask)

	return router
}
