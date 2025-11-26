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
