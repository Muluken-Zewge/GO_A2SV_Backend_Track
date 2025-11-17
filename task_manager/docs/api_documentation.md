# Task Manager API Documentation

A simple API for creating, reading, updating, and deleting tasks.

## 1. Overview and Base URL

All API endpoints are relative to the following base path.

- **Base API Path:** `/api/v1`
- **Full Local URL Example:** `http://localhost:8080/api/v1`

**Database:** MongoDB Atlas (Non-relational document store)

## 2. Authentication

This API is open and does not require authentication.

## 3. Data Model: `Task` Object

The primary resource object handled by the API has the following structure. Note that the **public `ID` field** is a custom identifier used for all API operations, separate from MongoDB's internal `_id`.

| Field         | Type     | Description                                                                                | Required on Create?            |
| :------------ | :------- | :----------------------------------------------------------------------------------------- | :----------------------------- | --- |
| **`ID`**      | `string` | The **unique identifier** for the task (server-generated custom ID, used for all lookups). | No                             |
| `Title`       | `string` | The name or title of the task.                                                             | **Yes**                        |
| `Description` | `string` | A detailed description of the task.                                                        | **Yes**                        |
| `DueDate`     | `string` | The date the task is due (ISO-8601/RFC3339 format).                                        | No (defaults to creation time) |
| `Status`      | `string` | The current status (e.g., "pending", "in-progress", "completed").                          | **Yes**                        |     |

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

## 4. Endpoints

### 4.1. Get All Tasks

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

### 4.2. Create a New Task

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

### 4.3. Get a Single Task

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

### 4.4. Update a Task

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

### 4.5. Delete a Task

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
