# go-serve-intro

A RESTful web server built in Go for managing a Twitter-like "Chirpy" social media platform. This project demonstrates web server basics including HTTP routing, database operations, authentication, and API design.

## Features

- User authentication with JWT tokens and refresh tokens
- CRUD operations for chirps (posts)
- User management and profile updates
- Content filtering (profanity filtering)
- File server with hit tracking
- Admin endpoints for metrics and management
- Webhook integration for user upgrades

## Tech Stack

- **Language**: Go 1.24.0
- **Database**: PostgreSQL (via `lib/pq`)
- **Authentication**: JWT tokens with Argon2id password hashing
- **Server**: Standard `net/http` package

## Getting Started

### Prerequisites

- Go 1.24.0 or later
- PostgreSQL database
- Environment variables configured in `./secrets/.env`

### Environment Variables

Required environment variables:
- `DB_URL`: PostgreSQL connection string
- `TOKEN_SECRET`: Secret key for JWT token signing
- `POLKA_KEY`: API key for Polka webhook authentication
- `PLATFORM`: Environment platform (e.g., "dev" for development)

### Running the Server

```bash
go run main.go
```

The server will start on port `:8080`.

---

## API Documentation

### Base URL

```
http://localhost:8080
```

### Authentication

Most endpoints require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer <JWT_TOKEN>
```

---

## Endpoints

### Health Check

#### `GET /api/healthz`

Check if the API is running.

**Response:**
- **Status Code**: `200 OK`
- **Content-Type**: `text/plain`
- **Body**: `OK`

---

### Chirps

#### `GET /api/chirps`

Get all chirps with optional filtering and sorting.

**Query Parameters:**
- `author_id` (optional): Filter chirps by user ID
- `sort` (optional): Sort order - `"asc"` (oldest first) or `"desc"` (newest first). Defaults to `"asc"`

**Response:**
- **Status Code**: `200 OK`
- **Content-Type**: `application/json`

**Response Body:**
```json
[
  {
    "id": "string",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "body": "string",
    "user_id": "string"
  }
]
```

**Example:**
```bash
GET /api/chirps?author_id=123&sort=desc
```

---

#### `GET /api/chirps/{id}`

Get a specific chirp by ID.

**Path Parameters:**
- `id`: Chirp ID

**Response:**
- **Status Code**: `200 OK` or `404 Not Found`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "id": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "body": "string",
  "user_id": "string"
}
```

**Error Response:**
```json
{
  "error": "Chirp not found"
}
```

---

#### `POST /api/chirps`

Create a new chirp. Requires authentication.

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "body": "string"
}
```

**Validation:**
- Body must be 140 characters or less
- Profanity filtering: words like "kerfuffle", "sharbert", "fornax" are replaced with `****`

**Response:**
- **Status Code**: `201 Created` or `400 Bad Request` or `401 Unauthorized`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "id": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "body": "cleaned body string",
  "user_id": "string"
}
```

**Error Responses:**
```json
{
  "error": "Chirp is too long"
}
```
```json
{
  "error": "Content-Type must be application/json"
}
```

---

#### `PUT /api/chirps`

Update an existing chirp. Requires authentication.

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "id": "string",
  "body": "string"
}
```

**Validation:**
- Body must be 140 characters or less
- Chirp must exist
- Profanity filtering applied

**Response:**
- **Status Code**: `200 OK` or `400 Bad Request` or `401 Unauthorized` or `404 Not Found`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "id": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "body": "updated cleaned body string",
  "user_id": "string"
}
```

---

#### `DELETE /api/chirps/{chirp_id}`

Delete a chirp. Requires authentication. Users can only delete their own chirps.

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`

**Path Parameters:**
- `chirp_id`: Chirp ID to delete

**Response:**
- **Status Code**: `204 No Content` or `401 Unauthorized` or `403 Forbidden` or `404 Not Found`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "message": "Chirp deleted successfully"
}
```

**Error Responses:**
```json
{
  "error": "You are not authorized to delete this chirp"
}
```

---

#### `POST /api/validate_chirp`

Validate a chirp body without creating it. Useful for client-side validation.

**Headers:**
- `Content-Type: application/json`

**Request Body:**
```json
{
  "body": "string"
}
```

**Response:**
- **Status Code**: `200 OK` or `400 Bad Request`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "body": "cleaned body string"
}
```

---

### Users

#### `POST /api/users`

Create a new user account.

**Headers:**
- `Content-Type: application/json`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**
- **Status Code**: `201 Created` or `400 Bad Request` or `500 Internal Server Error`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "id": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

**Note**: Password is hashed using Argon2id before storage.

---

#### `PUT /api/users`

Update user email and/or password. Requires authentication.

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "email": "newemail@example.com",
  "password": "newpassword"
}
```

At least one of `email` or `password` must be provided.

**Response:**
- **Status Code**: `200 OK` or `400 Bad Request` or `401 Unauthorized` or `404 Not Found`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "id": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "updated@example.com",
  "is_chirpy_red": false
}
```

---

### Authentication

#### `POST /api/login`

Authenticate a user and receive JWT and refresh tokens.

**Headers:**
- `Content-Type: application/json`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password",
  "expires_in_seconds": 3600
}
```

**Query Parameters:**
- `expires_in_seconds` (optional): Token expiration time in seconds. Defaults to 3600 (1 hour).

**Response:**
- **Status Code**: `200 OK` or `401 Unauthorized`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "id": "string",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "email": "user@example.com",
  "token": "JWT_TOKEN",
  "refresh_token": "REFRESH_TOKEN",
  "is_chirpy_red": false
}
```

**Error Response:**
```json
{
  "error": "Invalid email or password"
}
```

---

#### `POST /api/refresh`

Refresh an access token using a refresh token.

**Headers:**
- `Authorization: Bearer <REFRESH_TOKEN>`
- `Content-Type: application/json`

**Response:**
- **Status Code**: `200 OK` or `401 Unauthorized`
- **Content-Type**: `application/json`

**Success Response:**
```json
{
  "token": "NEW_JWT_TOKEN"
}
```

**Error Responses:**
```json
{
  "error": "Refresh token expired"
}
```
```json
{
  "error": "Refresh token revoked"
}
```

---

#### `POST /api/revoke`

Revoke a refresh token.

**Headers:**
- `Authorization: Bearer <REFRESH_TOKEN>`
- `Content-Type: application/json`

**Response:**
- **Status Code**: `204 No Content` or `401 Unauthorized` or `404 Not Found`
- **Content-Type**: `text/plain`
- **Body**: Empty

---

### Webhooks

#### `POST /api/polka/webhooks`

Webhook endpoint for Polka integration to upgrade users to Chirpy Red status.

**Headers:**
- `Authorization: ApiKey <POLKA_KEY>`
- `Content-Type: application/json`

**Request Body:**
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "string"
  }
}
```

**Response:**
- **Status Code**: `204 No Content` (success) or `400 Bad Request` or `401 Unauthorized` or `404 Not Found`
- **Content-Type**: `text/plain` or `application/json`

**Note**: Only `user.upgraded` events are processed. Other events return `204 No Content` with a message.

---

### Admin

#### `GET /admin/metrics`

Get server metrics (file server hit count).

**Response:**
- **Status Code**: `200 OK`
- **Content-Type**: `text/html`

**Response Body:**
```html
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited X times!</p>
  </body>
</html>
```

---

#### `POST /admin/reset`

Reset server metrics and delete all users. **Only available in dev environment.**

**Query Parameters:**
- Requires `PLATFORM=dev` environment variable

**Response:**
- **Status Code**: `200 OK` or `403 Forbidden`
- **Content-Type**: `text/plain`
- **Body**: `OK` or `Forbidden`

---

### Static Files

#### `GET /app/*`

Serve static files from the root directory.

**Note**: File server hits are tracked and displayed in `/admin/metrics`.

---

## Error Handling

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

### Common HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `204 No Content`: Request successful, no content to return
- `400 Bad Request`: Invalid request format or validation error
- `401 Unauthorized`: Authentication required or invalid
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

## Content Filtering

The API automatically filters profanity in chirp bodies. The following words are replaced with `****`:
- kerfuffle
- sharbert
- fornax

Filtering is case-insensitive and applies to partial word matches.

---

## Notes

- Chirps have a maximum length of 140 characters
- JWT tokens are used for authentication on protected endpoints
- Refresh tokens allow obtaining new access tokens without re-authentication
- User passwords are hashed using Argon2id before storage
- The server uses PostgreSQL for data persistence
- All timestamps are in UTC format
