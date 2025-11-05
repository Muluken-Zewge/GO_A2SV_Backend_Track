package main

import (
	"librarymanagement/controllers"
	"librarymanagement/models"
	"librarymanagement/services"
)

func main() {

	// Initialize the Service Layer (The Library State)
	libraryService := services.NewLibrary()

	controller := controllers.NewController(libraryService)

	// create a memeber
	memeberId := controller.NextMemberID
	controller.NextMemberID++

	libraryService.Members[memeberId] = models.Member{
		ID:            memeberId,
		Name:          "Baka",
		BorrowedBooks: []models.Book{},
	}

	controller.Start()
}
