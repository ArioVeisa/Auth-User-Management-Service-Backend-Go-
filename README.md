# Auth User Management Service (Go Backend)

This project is a backend service for authentication and user management, built with Go, PostgreSQL, and Redis. It provides features such as user registration, login with JWT, role-based access control, audit logging, and email integration for account-related workflows.

## High-Level Architecture

- **Language & Runtime**: Go.
- **Database**: PostgreSQL (stores users, roles, and audit logs).
- **Cache / Rate limiting / Session-related data**: Redis.
- **Authentication**: JWT (JSON Web Token) with a configurable signing key.
- **Email**: SMTP-based email sending (e.g., Gmail SMTP).
- **Containerization**: Docker + Docker Compose to run the app, PostgreSQL, and Redis together.

Main directories:

- `cmd/server` – application entrypoint (starts the HTTP server).
- `internal/config` – application configuration loading and helpers.
- `internal/database` – database and Redis connection setup.
- `internal/handlers` – HTTP handlers (auth, user, role, audit, health).
- `internal/middleware` – authentication and rate limiting middleware.
- `internal/models` – domain and persistence models.
- `internal/repository` – all database access logic.
- `internal/services` – business logic for auth, user, role, audit, and email.
- `internal/utils` – helper utilities (password hashing, JWT, validation, etc.).
- `migrations` – SQL migrations to initialize the PostgreSQL schema.

## Features

- **Authentication & Authorization**
  - User registration.
  - User login that returns JWT access tokens.
  - JWT-based middleware to protect private endpoints.
  - Role-based access control via roles and permissions.

- **User Management**
  - CRUD operations on users (create, read, update, delete), depending on the caller’s role.
  - Secure password storage using hashing.

- **Audit Logging**
  - Records important actions (such as login, user changes) into an audit log table.

- **Email Integration**
  - Sends application emails via SMTP (e.g., for verification or password reset).

- **Health Check**
  - Health endpoint to check that the app and its dependencies are running.

## Running with Docker Compose

Prerequisites:

- Docker
- Docker Compose

Steps:

1. **(Optional but recommended) Prepare environment file**

   ```bash
   cp .env.example .env
   ```

   Then edit `.env` to set values like `JWT_SIGNING_KEY`, SMTP credentials, etc.

2. **Build and start the services**

   From the project root:

   ```bash
   docker-compose up --build
   ```

   This will start:

   - `app` – the Go backend (exposed by default at `http://localhost:8080`).
   - `postgres` – PostgreSQL database (listening on container port `5432`, mapped to host port `5433` by default).
   - `redis` – Redis instance (listening on host port `6379` by default).

3. **Accessing the API**

   - Base URL: `http://localhost:8080`
   - Example endpoints (actual paths may vary with implementation):
     - `POST /auth/register` – register a new user.
     - `POST /auth/login` – authenticate a user and return a JWT.
     - `GET /users` – list users (requires authentication and proper role).
     - `GET /health` – simple health check endpoint.

## Running Locally without Docker

You can also run the app directly with your local Go toolchain.

1. Ensure you have local PostgreSQL and Redis instances running.
2. Set the required environment variables, for example:

   ```bash
   export APP_ENV=dev
   export PORT=8080
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable"
   export REDIS_URL="redis://localhost:6379"
   export JWT_SIGNING_KEY="your-super-secret-key-change-in-production"
   export SMTP_HOST="smtp.gmail.com"
   export SMTP_PORT="587"
   export SMTP_USER="your_email@gmail.com"
   export SMTP_PASSWORD="your_app_password"
   ```

3. Apply the SQL migration in `migrations/001_init_schema.up.sql` using your preferred method (e.g., `psql` or a migration tool).
4. Run the server:

   ```bash
   go run ./cmd/server
   ```

The server will listen on `http://localhost:8080` by default.

## Important Environment Variables

- `APP_ENV` – application environment (`dev`, `prod`, etc.).
- `PORT` – HTTP server port (default `8080`).
- `DATABASE_URL` – PostgreSQL connection URL.
- `REDIS_URL` – Redis connection URL.
- `JWT_SIGNING_KEY` – signing key used for JWT tokens (must be changed for production).
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD` – SMTP configuration for sending emails.

## Build and Deployment

- **Build binary locally**:

  ```bash
  go build -o auth-service ./cmd/server
  ```

- **Build Docker image**:

  ```bash
  docker build -t auth-service:latest .
  ```

You can then deploy the image to your container platform of choice.

## Git Workflow (Example)

Typical flow after making changes:

```bash
git status
git add .
git commit -m "Update docs and configuration"
git push origin main
```

Adjust the branch name and remote as needed.