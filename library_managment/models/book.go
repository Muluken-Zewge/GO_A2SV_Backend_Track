package models

type Book struct {
	ID     int
	Title  string
	Author string
	Status string
}

// ReservationRequest is the job we put on the queue
type ReservationRequest struct {
	BookID    int
	MemberID  int
	ReplyChan chan error
}
