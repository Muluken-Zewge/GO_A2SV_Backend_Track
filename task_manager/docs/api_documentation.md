# Task Manager API Documentation

A simple API for creating, reading, updating, and deleting tasks, secured with JWT-based authentication and role-based authorization.

## 1. Overview and Base URL

All API endpoints are relative to the following base path.

- **Base API Path:** `/api/v1`
- **Full Local URL Example:** `http://localhost:8080/api/v1`

**Database:** MongoDB Atlas (Non-relational document store)

## 2. Authentication and Authorization (Security)🔐

This API requires a valid JSON Web Token (JWT) for access to most endpoints. Access is further restricted based on the user's role.

### 2.1. User Roles

| Role  | Numeric Value | Permissions                                                    |
| :---- | :------------ | :------------------------------------------------------------- |
| User  | 0             | Read-Only access to all tasks (GET /tasks).                    |
| Admin | 1             | Full CRUD access to all tasks, plus user management functions. |

### 2.2. Obtaining and Using the JWT

1. Login: Send a POST request to /api/v1/user/login.

2. Token Retrieval: The successful response will contain a long string, the JWT.

3. Usage: For all protected endpoints (listed below), include the JWT in the Authorization Header of your request.
   | Header | Format | Example |
   | :--- | :--- | :--- |
   | Authorization | Bearer <JWT_TOKEN> | Bearer eyJhbGciOiJIUzI1NiIsInR5c... |

### 2.3. Common Error Responses

| Status Code | Description  | Meaning                                                                                                                    |
| :---------- | :----------- | :------------------------------------------------------------------------------------------------------------------------- |
| 401         | Unauthorized | Authentication failure (e.g., Missing token, expired token, wrong password). You are not logged in.                        |
| 403         | Forbidden    | Authorization failure (e.g., A User role trying to access an Admin-only endpoint). You are logged in, but lack permission. |

## 3. Data Models🏗️

### 3.1. User Object (Registration/Login)

| Field     | Type   | Description                    |
| :-------- | :----- | :----------------------------- |
| id        | string | Unique identifier (UUID).      |
| user_name | string | Unique username.               |
| role      | int    | User's role (0=User, 1=Admin). |

### 3.2. Task Object

The primary resource object handled by the API has the following structure. Note that the **public `ID` field** is a custom identifier used for all API operations, separate from MongoDB's internal `_id`.

| Field       | Type   | Description                                         | Required on Create? |
| :---------- | :----- | :-------------------------------------------------- | :------------------ |
| ID          | string | The unique identifier for the task.                 | No                  |
| Title       | string | The name or title of the task.                      | Yes                 |
| Description | string | A detailed description of the task.                 | Yes                 |
| Due Date    | string | The date the task is due (ISO-8601/RFC3339 format). | No                  |
| Status      | string | The current status (e.g., "pending", "completed").  | Yes                 |

**Example `Task` Object:**

```json
{
  "ID": "1",
  "Title": "Finish API documentation",
  "Description": "Write the complete documentation for all endpoints.",
  "DueDate": "2025-11-12T14:30:00Z",
  "Status": "in-progress"
}
```

## 4. User Endpoints 👤

All paths are relative to /api/v1/user.

### 4.1. Register User

Creates a new user account. The very first user registered is automatically assigned the Admin role. Subsequent users are assigned the User role.

| Method | Path           | Access |
| :----- | :------------- | :----- |
| POST   | /user/register | Public |

Request Body:

```json
{
  "user_name": "new_user_name",
  "password": "strongpassword123"
}
```

Success Response (201 Created):

```json
{
  "message": "user created successfully"
}
```

### 4.2. Authenticate User (Login)

Authenticates the user and returns a JWT token.

| Method | Path        | Access |
| :----- | :---------- | :----- |
| POST   | /user/login | Public |

Request Body:

```json
{
  "user_name": "admin_user",
  "password": "strongpassword123"
}
```

Success Response (200 OK):

```json
{
  "message": "Login successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDE5NzAwMDAsInJvbGUiOjF9..."
}
```

### 4.3. Promote User Role

Allows an Admin to promote any existing user to the Admin role.

| Method | Path         | Access     |
| :----- | :----------- | :--------- |
| PATCH  | /:id/promote | Admin Only |

Success Response (200 OK):

```json
{
  "message": "user status updated successfully",
  "updatedUser": {
    "ID": "a65c92...",
    "UserName": "promoted_user",
    "Role": 1
  }
}
```

## 5. Task Endpoints📝

All paths are relative to /api/v1/tasks.

### 5.1. Get All Tasks

Retrieves a list of all tasks.

| Detail     | Value    |
| ---------- | -------- |
| **Method** | GET      |
| **Path**   | `/tasks` |

Success Response (200 OK):

```json
{
  "tasks": [
    {
      "ID": "1",
      "Title": "Finish API documentation",
      "Description": "Write the complete documentation...",
      "DueDate": "2025-11-12T14:30:00Z",
      "Status": "in-progress"
    }
  ]
}
```

### 5.2. Create a New Task

Creates a new task.

| Detail     | Value    |
| ---------- | -------- |
| **Method** | POST     |
| **Path**   | `/tasks` |

Request Body (Required Fields: Title, Description, Status):

```json
{
  "Title": "Deploy to production",
  "Status": "pending",
  "Description": "Push the final code to the server."
}
```

Success Response (201 Created):

```json
{
  "message": "Task created successfully",
  "task": {
    "ID": "3",
    "Title": "Deploy to production",
    "Description": "Push the final code to the server.",
    "DueDate": "2025-11-12T15:01:59Z",
    "Status": "pending"
  }
}
```

Error Response (400 Bad Request):

```json
{
  "error": "title, description and status are required fields"
}
```

### 5.3. Get a Single Task

Retrieves a single task by its unique ID.

| Detail     | Value        |
| ---------- | ------------ |
| **Method** | GET          |
| **Path**   | `/tasks/:id` |

Success Response (200 OK):

```json
{
  "ID": "2",
  "Title": "Test the API",
  "Description": "Use Postman to test all endpoints.",
  "DueDate": "2025-11-13T10:00:00Z",
  "Status": "pending"
}
```

Error Response (404 Not Found):

```json
{
  "error": "task not found"
}
```

### 5.4. Update a Task

Updates an existing task. Only the fields provided in the JSON body will be updated. All fields are optional.

| Detail     | Value        |
| ---------- | ------------ |
| **Method** | PUT          |
| **Path**   | `/tasks/:id` |

Request Body (All fields optional for update):

```json
{
  "Status": "completed",
  "Description": "All endpoints tested and working."
}
```

Success Response (200 OK):

```json
{
  "message": "Task updated successfully",
  "Updated Task": {
    "ID": "2",
    "Title": "Test the API",
    "Description": "All endpoints tested and working.",
    "DueDate": "2025-11-13T10:00:00Z",
    "Status": "completed"
  }
}
```

Error Response (404 Not Found):

```json
{
  "error": "task not found"
}
```

### 5.5. Delete a Task

Deletes a task by its unique ID.

| Detail     | Value        |
| ---------- | ------------ |
| **Method** | DELETE       |
| **Path**   | `/tasks/:id` |

Success Response (200 OK):

```json
{
  "message": "Task deleted"
}
```

Error Response (404 Not Found):

```json
{
  "error": "task not found"
}
```

## 🧪 Testing Guide

This project uses a layered testing strategy to ensure reliability across the domain, usecases, and delivery layers. We use the **Testify** library for assertions and suites, and **Mockery** for dependency injection.

---

## 1. Prerequisites

Before running tests, ensure you have the following installed:

- **Go:** 1.21 or higher
- **Mockery:** To regenerate mocks if you change interfaces

Install Mockery using:

```bash
go install github.com/vektra/mockery/v2@latest
```

## 2. Test Setup & Organization

Tests are co-located with the source code using the `_test.go` suffix.

- **Unit Tests:**  
  Located in `Usecases/` and `Infrastructure/`. These test business logic in isolation.

- **Controller Tests:**  
  Located in `Delivery/controllers/`. These test the logic of handling requests and mapping errors.

- **Router / Integration Tests:**  
  Located in `Delivery/router/`. These verify middleware, routing, and end-to-end HTTP flows.

Environment Variables
The router and middleware tests require a JWT_SECRET. This is handled automatically within the test suites using `os.Setenv("JWT_SECRET", "test_secret")`, so no manual .env setup is required for testing.

## 3. Running Tests

### Run all tests

To run every test in the project:

```bash
go test ./...
```

Run tests with verbose output

Useful for seeing individual test names and fmt.Printf output:

```bash
go test -v ./...
```

Check Test Coverage
To see which lines of code are covered by tests:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 4. Mocking with Mockery

We use mocks to isolate the layer we are testing. If you modify any Interface (e.g., in UserRepository or TaskUsecase), you must update the mocks:

```bash
mockery --all
```

This will update all files in the `mocks/` directory based on your current interfaces.

## 5. Understanding Test Output

✅ Successful Test
When a test passes, you will see a simple `PASS` notification:

```plaintext
=== RUN   TestTaskController_GetTasks_Success
--- PASS: TestTaskController_GetTasks_Success (0.00s)
PASS
ok      taskmanager/Delivery/controllers  0.123s

❌ Failed Test
When a test fails, Testify provides a "Diff" showing what was expected vs. what was actually received:
```

```plaintext
=== RUN   TestAuthMiddleware_Fail_NoHeader
    middleware_test.go:85:
        Error Trace:    middleware_test.go:85
        Error:          Not equal:
                        expected: 401
                        actual:   200
        Test:           TestAuthMiddleware_Fail_NoHeader
--- FAIL: TestAuthMiddleware_Fail_NoHeader (0.00s)
FAIL
```

How to read this:

Error Trace: The exact file and line number where the failure occurred.

Expected vs Actual: Shows that the middleware was expected to return 401 Unauthorized but returned 200 OK instead (meaning the middleware failed to abort the request).

## 6. Adding New Tests

When adding new features:

Define the Interface in the Domain layer.

Run `mockery --all` to generate the new mock.

Create a `_test.go` file in the relevant directory.

Use `httptest.NewRecorder()` for any logic involving Gin controllers or middleware.
