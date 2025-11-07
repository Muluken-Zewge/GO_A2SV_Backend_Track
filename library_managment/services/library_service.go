package services

import (
	"errors"
	"fmt"
	"librarymanagement/models"
	"sync"
	"time"
)

type Library struct {
	Books            map[int]models.Book
	Members          map[int]models.Member
	Mutex            sync.Mutex
	ReservationQueue chan *models.ReservationRequest
}

// constructor function to create a new library instance
func NewLibrary() *Library {
	return &Library{
		Books:   make(map[int]models.Book),
		Members: make(map[int]models.Member),
		Mutex:   sync.Mutex{},
		// create a buffered channel to hold 100 requests
		ReservationQueue: make(chan *models.ReservationRequest, 100),
	}
}

type LibraryManager interface {
	AddBook(book models.Book)
	RemoveBook(bookID int) error
	BorrowBook(bookID int, memberID int) error
	ReturnBook(bookID int, memberID int) error
	ListAvailableBooks() []models.Book
	ListBorrowedBooks(memberID int) []models.Book
	ReserveBook(bookID int, memberID int) error
}

func (l *Library) AddBook(book models.Book) {
	l.Books[book.ID] = book
}

func (l *Library) RemoveBook(bookID int) error {
	_, ok := l.Books[bookID]
	if !ok {
		return errors.New("INVALID BOOK ID")
	}

	delete(l.Books, bookID)
	return nil
}

func (l *Library) BorrowBook(bookID int, memberID int) error {
	book, ok := l.Books[bookID]
	if !ok {
		return errors.New("BOOK DOESN'T EXIST")
	}

	if book.Status != "Available" {
		return errors.New("BOOK IS UNAVAILABLE")
	}

	member, ok := l.Members[memberID]
	if !ok {
		return errors.New("MEMBER DOESN'T EXIST")
	}

	member.BorrowedBooks = append(member.BorrowedBooks, book)
	book.Status = "Borrowed"

	// put the updated structs back to the library modify the original
	l.Members[memberID] = member
	l.Books[bookID] = book
	return nil
}

func (l *Library) ReturnBook(bookID int, memberID int) error {
	book, ok := l.Books[bookID]
	if !ok {
		return errors.New("BOOK DOESN'T EXIST")
	}

	if book.Status != "Borrowed" {
		return errors.New("THE BOOK WAS NEVER BORROWED, YOU CAN ADD A BOOK WITH '1. Add a New Book'")
	}

	member, ok := l.Members[memberID]
	if !ok {
		return errors.New("MEMBER DOESN'T EXIST")
	}

	for i, b := range member.BorrowedBooks {
		if b.ID == book.ID {
			member.BorrowedBooks = append(member.BorrowedBooks[:i], member.BorrowedBooks[i+1:]...)
		}
	}

	book.Status = "Available"

	// put back to the library
	l.Members[memberID] = member
	l.Books[bookID] = book
	return nil

}

func (l *Library) ListAvailableBooks() []models.Book {
	availableBooks := make([]models.Book, 0, len(l.Books))

	for _, book := range l.Books {
		if book.Status == "Available" {
			availableBooks = append(availableBooks, book)
		}
	}

	return availableBooks
}

func (l *Library) ListBorrowedBooks(memberID int) []models.Book {
	if member, ok := l.Members[memberID]; ok {
		return member.BorrowedBooks
	}
	return []models.Book{}
}

// creates and sends the requests
func (l *Library) ReserveBook(bookID int, memberID int) error {
	// create a channel for the worker to reply to
	replyChan := make(chan error, 1)

	// Create the request
	request := &models.ReservationRequest{
		BookID:    bookID,
		MemberID:  memberID,
		ReplyChan: replyChan,
	}

	// Put the ticket on the queue
	l.ReservationQueue <- request

	// Wait for the worker to send a reply
	err := <-replyChan

	return err

}

func (l *Library) DoReservation(bookID int, memberID int) error {
	// LOCK the mutex at the beggining to make sure only one go routine can check/change the map at a time
	l.Mutex.Lock()

	// defer to guarantee the mutex is unlocked when the function returns
	defer l.Mutex.Unlock()

	book, ok := l.Books[bookID]
	if !ok {
		return errors.New("BOOK DOESN'T EXIST")
	}

	_, ok = l.Members[memberID]
	if !ok {
		return errors.New("MEMBER DOESN'T EXIST")
	}

	if book.Status == "Reserved" {
		return errors.New("BOOK IS ALREADY RESERVED")
	}

	if book.Status == "Available" {
		book.Status = "Reserved"
		l.Books[bookID] = book

		go l.handleReservationTimeout(bookID)

		return nil
	}

	return errors.New("BOOK IS NOT AVAILABLE")
}

func (l *Library) handleReservationTimeout(bookID int) {
	// wait for 5 seconds in the background
	time.Sleep(5 * time.Second)

	// lock the mutex to avoid race condition
	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	// re-fetch the book, as it might have changed
	book, ok := l.Books[bookID]
	if !ok {
		return
	}
	// check if the book is still in reserved status(not borrowed)
	if book.Status == "Reserved" {
		book.Status = "Available"
		l.Books[bookID] = book
		fmt.Printf("\n[AUTO-CANCEL] Reservation for Book ID %d expired.\n", bookID)
	}
}
