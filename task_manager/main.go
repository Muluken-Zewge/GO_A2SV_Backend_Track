package main

import "taskmanager/router"

func main() {
	r := router.SetupRouter()

	// run server
	r.Run(":8080")
}
