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

---

## ⚡ Concurrency Implementation

To handle multiple simultaneous reservation requests safely, the system uses a **Worker Pool** pattern combined with a **Mutex** for thread-safe state management.

This design decouples the user's _request_ from the actual _processing_ of the reservation, ensuring the system remains responsive and free from race conditions.

### 1. The Reservation Queue (Channels)

- **What it is:** A global, buffered channel (`ReservationQueue`) is added to the `Library` struct.
- **How it works:** When a user tries to reserve a book, the `ReserveBook` function does not perform the action. Instead, it creates a `ReservationRequest` job (a "ticket") and places it on this channel (the "queue").
- **Analogy:** This is like a waiter (the `ReserveBook` function) putting an order on a ticket spike (the `ReservationQueue`) for the kitchen to handle.

### 2. The Worker Pool (Goroutines)

- **What it is:** At startup (`main.go`), the application launches several "worker" goroutines (`concurrency.StartWorkers`).
- **How it works:** These workers run in the background and constantly watch the `ReservationQueue`. When a job appears, the first available worker grabs it and executes the _actual_ reservation logic (`DoReservation`).
- **Analogy:** These are the chefs (the workers) who grab the tickets from the spike and cook the food (process the request).

### 3. Preventing Race Conditions (Mutex)

- **What it is:** A `sync.Mutex` is added to the `Library` struct.
- **How it works:** The "worker" goroutines must **lock** the mutex _before_ they are allowed to read or write to the `Library.Books` or `Library.Members` maps. Once done, they **unlock** it.
- **Why:** This ensures that only one worker can modify the library's state at any given moment, preventing two users from reserving the same book at the exact same time.
- **Analogy:** This is the key to a single walk-in fridge. Only one chef (worker) can have the key at a time to get ingredients (access the maps).

### 4. Asynchronous Auto-Cancellation (Goroutine + Timer)

- **How it works:** After a worker successfully reserves a book, it launches _another_ new goroutine (`handleReservationTimeout`).
- This new goroutine simply sleeps for 5 seconds (`time.Sleep`).
- When it wakes up, it re-locks the mutex and checks if the book is _still_ "Reserved". If it is, it sets the status back to "Available".
- This happens entirely in the background, allowing the user and the workers to continue processing other requests.
