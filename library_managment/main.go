package main

import (
	"fmt"
	"librarymanagement/concurrency"
	"librarymanagement/controllers"
	"librarymanagement/models"
	"librarymanagement/services"
	"sync"
	"time"
)

func main() {

	// Initialize the Service Layer (The Library State)
	libraryService := services.NewLibrary()

	// start the workers(5)
	concurrency.Startworkers(5, libraryService.ReservationQueue, libraryService)

	controller := controllers.NewController(libraryService)

	// create a memeber
	memberId := controller.NextMemberID
	controller.NextMemberID++

	libraryService.Members[memberId] = models.Member{
		ID:            memberId,
		Name:          "Baka",
		BorrowedBooks: []models.Book{},
	}

	// create a book
	bookId := controller.NextBookID
	controller.NextBookID++
	libraryService.AddBook(models.Book{
		ID:     bookId,
		Title:  "Test Book",
		Author: "Test Author",
		Status: "Available",
	})

	// ---START OF CONCURRENCY TEST ---

	fmt.Println("=======================================")
	fmt.Println("🚀 STARTING CONCURRENCY TEST...")
	fmt.Printf("Simulating 3 users trying to reserve Book ID %d at the same time...\n", bookId)

	// A WaitGroup is needed to wait for all goroutines to finish
	var wg sync.WaitGroup
	numRequests := 3
	wg.Add(numRequests)

	for i := 1; i <= numRequests; i++ {
		// Launch a new goroutine for each "user"
		go func(userID int) {
			defer wg.Done() // Tell the WaitGroup this goroutine is done when it returns

			// Each goroutine tries to reserve the *same book*
			err := libraryService.ReserveBook(bookId, memberId)
			if err != nil {
				fmt.Printf("[User %d] FAILED to reserve: %s\n", userID, err)
			} else {
				fmt.Printf("[User %d] SUCCESS: Book reserved!\n", userID)
			}
		}(i)
	}

	// Wait here until all 3 goroutines have called wg.Done()
	wg.Wait()
	fmt.Println("...Concurrency test finished.")
	fmt.Println("---")

	// --- TEST AUTO-CANCELLATION ---

	fmt.Println("Waiting 6 seconds to test auto-cancellation...")
	time.Sleep(6 * time.Second)
	// The auto-cancel message should appear while we sleep.

	// Check the book status after the timeout
	// We must use the mutex to read safely!
	libraryService.Mutex.Lock()
	finalBookStatus := libraryService.Books[bookId].Status
	libraryService.Mutex.Unlock()

	fmt.Printf("Auto-cancel test complete. Final book status: %s (should be 'Available')\n", finalBookStatus)

	// start the normal interactive app
	controller.Start()
}
