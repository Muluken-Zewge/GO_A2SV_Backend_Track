package controllers

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"librarymanagement/models"
	"librarymanagement/services"
)

type LibraryController struct {
	Service      services.LibraryManager
	Scanner      *bufio.Scanner
	NextBookID   int
	NextMemberID int
}

func NewController(svc services.LibraryManager) *LibraryController {
	return &LibraryController{
		Service:      svc,
		Scanner:      bufio.NewScanner(os.Stdin),
		NextBookID:   1,
		NextMemberID: 1,
	}
}

// runs the main interactive command loop
func (c *LibraryController) Start() {
	c.printWelcome() // print welcome message

	for {
		c.printMenu() // print available commands

		fmt.Println("Enter option number: ")
		if !c.Scanner.Scan() {
			fmt.Println("\nExiting Library manager...")
			break
		}

		input := strings.TrimSpace(c.Scanner.Text())
		if input == "" {
			fmt.Println("Please enter an option number!")
			continue
		}

		options, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Error: Invalid input. Please enter a number from a menu.")
			continue
		}

		if !c.handleMenuOption(options) {
			break // exit the loop if handle menu option returns false
		}

	}
}

// routes the choosen number to the appropriate function
func (c *LibraryController) handleMenuOption(option int) bool {
	switch option {
	case 1:
		c.handleAddBook()
	case 2:
		c.handleRemoveBook()
	case 3:
		c.handleBorrowBook()
	case 4:
		c.handleReturnBooK()
	case 5:
		c.handleListAvailable()
	case 6:
		c.handleListBorrowed()
	case 7:
		fmt.Println("Goodbye!")
		return false // signal to exis the loop
	default:
		fmt.Println("Error: Options not recognized. Please choose a valid number.")
	}
	return true // continue the loop
}

// --- input helpers ---

// prints the prompt and returns the user's input line
func (c *LibraryController) promptForInput(prompt string) string {
	fmt.Printf("%s: ", prompt)
	if c.Scanner.Scan() {
		return strings.TrimSpace(c.Scanner.Text())
	}
	return ""
}

// prints the prompts and returns the parsed integer(id)
func (c *LibraryController) promptForInt(prompt string) (int, error) {
	input := c.promptForInput(prompt)
	return strconv.Atoi(input)
}

// --- COMMAND IMPLEMENTAION ---

func (c *LibraryController) handleAddBook() {
	fmt.Println("\n--- And Book ---")
	title := c.promptForInput("Enter book Title")
	author := c.promptForInput(("Enter book Author"))
	status := "Available"
	bookID := c.NextBookID
	c.NextBookID++

	book := models.Book{
		ID:     bookID,
		Title:  title,
		Author: author,
		Status: status,
	}

	c.Service.AddBook(book)
	fmt.Printf("Success: Added book '%s' with ID %d.\n", book.Title, bookID)
}

func (c *LibraryController) handleRemoveBook() {
	fmt.Println("\n--- Remove Book ---")
	id, err := c.promptForInt("Enter Book ID to remove")
	if err != nil {
		fmt.Println("Error: Invalid ID.")
		return
	}
	if err = c.Service.RemoveBook(id); err != nil {
		fmt.Println("Error: Invalid ID.")
		return
	}
	fmt.Printf("Success: Book ID %d removed.\n", id)
}

func (c *LibraryController) handleBorrowBook() {
	fmt.Println("\n--- Borrow Book ---")
	bookID, errB := c.promptForInt("Enter Book ID to borrow")
	memberID, errM := c.promptForInt("Enter Member ID")

	if errB != nil || errM != nil {
		fmt.Println("Error: Both IDs must be valid integers.")
		return
	}

	if err := c.Service.BorrowBook(bookID, memberID); err != nil {
		fmt.Printf("Error borrowing book: %v\n", err)
	} else {
		fmt.Printf("Success: Book ID %d borrowed by Member ID %d.\n", bookID, memberID)
	}
}

func (c *LibraryController) handleReturnBooK() {
	fmt.Println("\n--- Return Book ---")
	bookID, errB := c.promptForInt("Enter Book ID to return")
	memberID, errM := c.promptForInt("Enter Member ID")

	if errB != nil || errM != nil {
		fmt.Println("Error: Both IDs must be valid integers.")
		return
	}

	if err := c.Service.ReturnBook(bookID, memberID); err != nil {
		fmt.Printf("Error returning book: %v\n", err)
	} else {
		fmt.Printf("Success: Book ID %d returned by Member ID %d.\n", bookID, memberID)
	}
}

func (c *LibraryController) handleListAvailable() {
	books := c.Service.ListAvailableBooks()
	if len(books) == 0 {
		fmt.Println("No books are currently available.")
		return
	}
	fmt.Println("\n--- Available Books ---")
	for _, book := range books {
		fmt.Printf("ID: %d | Title: %s | Author: %s | Status: %s\n", book.ID, book.Title, book.Author, book.Status)
	}
	fmt.Println("-----------------------")
}

func (c *LibraryController) handleListBorrowed() {
	fmt.Println("\n--- List Borrowed Books ---")
	memberID, err := c.promptForInt("Enter Member ID")
	if err != nil {
		fmt.Println("Error: Member ID must be an integer.")
		return
	}

	books := c.Service.ListBorrowedBooks(memberID)
	if len(books) == 0 {
		fmt.Printf("Member ID %d has no borrowed books.\n", memberID)
		return
	}
	fmt.Printf("\n--- Books Borrowed by Member ID %d ---\n", memberID)
	for _, book := range books {
		fmt.Printf("ID: %d | Title: %s | Author: %s\n", book.ID, book.Title, book.Author)
	}
	fmt.Println("---------------------------------------")
}

// --- WELCOME/MENU HELPERS ---

func (c *LibraryController) printWelcome() {
	println("=========================================")
	println("📚 WELCOME TO THE LIBRARY MANAGER CLI 📚")
	println("=========================================")
}

func (c *LibraryController) printMenu() {
	fmt.Println("---------------------")
	fmt.Println("Available Commands:")
	fmt.Println("1. Add a New Book")
	fmt.Println("2. Remove Book by ID")
	fmt.Println("3. Borrow Book")
	fmt.Println("4. Return Book")
	fmt.Println("5. List Available Books")
	fmt.Println("6. List Books Borrowed by Member")
	fmt.Println("7. Exit")
	fmt.Println("---------------------")

}
