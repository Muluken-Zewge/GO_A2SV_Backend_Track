package concurrency

import (
	"fmt"
	"librarymanagement/models"
	"librarymanagement/services"
)

// laounches the goroutine workers to process reservation queue
func Startworkers(workerCount int, queue <-chan *models.ReservationRequest, lib *services.Library) {
	for i := 0; i <= workerCount; i++ {
		go func() {
			for request := range queue {
				// When a request appears, do the work
				err := lib.DoReservation(request.BookID, request.MemberID)

				// Send the result (nil or an error) back to the original caller
				request.ReplyChan <- err
			}
		}()
	}
	fmt.Printf("Started %d reservation workers.\n", workerCount)
}
