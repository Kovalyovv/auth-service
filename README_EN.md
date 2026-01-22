# Authentication Service (auth-service)

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This service is the central authentication authority for the project. It handles user registration, login, and the issuance of JWT access and refresh tokens. It provides an HTTP API for clients and a gRPC interface for other backend services to verify user tokens.

## Project Architecture

This project consists of three core microservices that work together:

-   **`auth-service` (This Service):** Manages all user-related concerns, including identity and token-based authentication.
-   **[Chat Service](../chat-app/README.md):** A real-time WebSocket-based messaging service that relies on `auth-service` to authenticate users.
-   **[Media Processor](../media-processor/README.md):** An asynchronous service for handling file uploads, which also uses `auth-service` for validating user permissions.

## Core Features

-   **User Management:** Secure user registration with password hashing and login via email and password.
-   **Token-Based Authentication:** Generates JWT access tokens (short-lived) and refresh tokens (long-lived) upon successful login.
-   **Token Refresh:** Allows clients to obtain a new access token using a valid refresh token without needing to re-enter credentials.
-   **gRPC for Internal Verification:** Exposes a high-performance gRPC endpoint for other services to quickly and securely verify access tokens.

## Technology Stack

-   **Language:** Go
-   **API Frameworks:**
    -   [Gin](https://github.com/gin-gonic/gin) for the HTTP server.
    -   [gRPC](https://grpc.io/) for high-performance inter-service communication.
-   **Database:** PostgreSQL, accessed via the [pgx/v5](https://github.com/jackc/pgx) driver.
-   **Authentication:** [jwt-go](https://github.com/golang-jwt/jwt) for creating and validating JSON Web Tokens.
-   **Logging:** `slog` (standard library) for structured logging.
-   **Testing:** [testify](https://github.com/stretchr/testify) for assertions and [testcontainers-go](https://github.com/testcontainers/testcontainers-go) for integration testing.

## API Endpoints

### HTTP API

All endpoints are prefixed with `/auth`.

| Method | Endpoint      | Description                                               |
| :----- | :------------ | :-------------------------------------------------------- |
| `POST` | `/register`   | Creates a new user account.                               |
| `POST` | `/login`      | Authenticates a user and returns an access/refresh token pair. |
| `POST` | `/refresh`    | Issues a new token pair using a valid refresh token.      |

### gRPC API

The service exposes a gRPC server for internal use.

| Service       | RPC Method    | Description                                       |
| :------------ | :------------ | :------------------------------------------------ |
| `AuthService` | `VerifyToken` | Verifies an access token and returns the user ID. |

## How to Run

### Recommended Method: Docker Compose

The easiest and recommended way to run the entire project is with Docker Compose. This will start this service, its database, and all other related services.

1.  Navigate to the project's root directory.
2.  Run the following command:
    ```bash
    docker compose up --build
    ```

### Local Development (Alternative)

You can run the service directly on your machine for development.

1.  **Prerequisites:**
    -   Go 1.24+ installed.
    -   A running PostgreSQL instance.

2.  **Configure Environment:**
    Create a `.env` file in the `auth-service` directory:
    ```env
    # Port for the HTTP server
    HTTP_PORT=8001

    # Port for the gRPC server
    GRPC_PORT=50001

    # Connection string for your PostgreSQL database
    DATABASE_URL=postgres://user:password@localhost:5432/auth_db?sslmode=disable

    # A strong, unique secret for signing JWTs
    JWT_SECRET=your-super-secret-key
    ```

3.  **Run the service:**
    ```bash
    go run ./cmd/auth/main.go
    ```
