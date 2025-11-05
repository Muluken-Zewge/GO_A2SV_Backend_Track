# 📚 Library Management System CLI

This is a simple, interactive terminal (CLI) application for managing library books and members, written in Go.

The application runs an interactive menu loop, allowing a user to manage the library's state.

---

## 🚀 How to Run

1.  Ensure you have **Go (1.21+)** installed on your system.
2.  Navigate to the project's root directory in your terminal.
3.  Run the application:

    ```bash
    go run main.go
    ```

---

## ⚙️ How to Use (Main Menu)

Once the application is running, you will be greeted with a welcome message and the main command menu.

Enter the number corresponding to the action you wish to perform and press **Enter**. Follow the on-screen prompts for each command.

### Menu Options

- **1. Add a New Book**

  - Prompts you to enter a **Title** and **Author**.
  - The application automatically assigns a new, unique **Book ID**.

- **2. Remove Book by ID**

  - Prompts you to enter the **Book ID** of the book you wish to remove.

- **3. Borrow Book**

  - Prompts for the **Book ID** to be borrowed and the **Member ID** of the borrower.
  - Updates the book's status to "Borrowed" and adds it to the member's list.

- **4. Return Book**

  - Prompts for the **Book ID** being returned and the **Member ID** of the member.
  - Updates the book's status to "Available" and removes it from the member's list.

- **5. List Available Books**

  - Displays a list of all books in the library with the status "Available".

- **6. List Books Borrowed by Member**

  - Prompts for a **Member ID**.
  - Displays a list of all books currently borrowed by that member.

- **7. Exit**
  - Closes the application.

---

## 📁 Project Structure

```
librarymanagement/
├── go.mod
├── main.go # Main entry point: initializes services and controllers
├── DOCUMENTATION.md # This file
├── controllers/
│ └── library_controller.go # Handles all terminal I/O (menus, prompts)
├── models/
│ ├── book.go # Defines the Book struct
│ └── member.go # Defines the Member struct
└── services/
└── library_service.go # Contains all business logic (Library struct, methods)
```
