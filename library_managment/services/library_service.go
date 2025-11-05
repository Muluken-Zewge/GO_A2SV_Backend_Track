package services

import (
	"errors"
	"librarymanagement/models"
)

type Library struct {
	Books   map[int]models.Book
	Members map[int]models.Member
}

// constructor function to create a new library instance
func NewLibrary() *Library {
	return &Library{
		Books:   make(map[int]models.Book),
		Members: make(map[int]models.Member),
	}
}

type LibraryManager interface {
	AddBook(book models.Book)
	RemoveBook(bookID int) error
	BorrowBook(bookID int, memberID int) error
	ReturnBook(bookID int, memberID int) error
	ListAvailableBooks() []models.Book
	ListBorrowedBooks(memberID int) []models.Book
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
