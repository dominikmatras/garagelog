# GarageLog

GarageLog is a work-in-progress REST API for managing vehicle records and user accounts. The repository currently contains a Go backend and a PostgreSQL development setup. There is no frontend yet.

## Current functionality

- User registration with bcrypt password hashing and unique email addresses.
- Login with an HS256 JWT containing the user ID and role, valid for 24 hours.
- Authenticated access to the current user's profile.
- An admin-only endpoint for listing users.
- Authenticated vehicle creation, listing, lookup, replacement, and deletion.
- A public health endpoint.

New accounts receive the `user` role in the database. The `admin` role is required to list users. Password hashes are excluded from JSON responses.

**Vehicle records are shared across all authenticated users.** There is no ownership relationship between users and vehicles; any authenticated user can read, update, or delete any vehicle.

## Stack

| Component | Implementation |
| --- | --- |
| Language | Go, with `go 1.26.0` declared in `backend/go.mod` |
| HTTP framework | Gin |
| Database | PostgreSQL 17, provided by Docker Compose |
| Database access | pgx connection pool and handwritten SQL |
| Authentication | golang-jwt/jwt v5 and bcrypt |
| Schema changes | Versioned SQL up/down migration files |

## Run locally

Prerequisites: a compatible Go toolchain, Docker with Docker Compose, and a PostgreSQL migration runner such as the `golang-migrate` CLI with PostgreSQL support (`migrate`). The commands below use PowerShell and start from the repository root.

### 1. Start PostgreSQL

```powershell
cd backend
docker compose up -d postgres
docker compose ps
```

Wait for the database to become healthy. Compose exposes PostgreSQL on port `5432` and stores its data in the `postgres_data` volume. It starts only the database; the API runs separately.

### 2. Configure the API

```powershell
$env:DATABASE_URL = "postgres://garagelog:garagelog@localhost:5432/garagelog?sslmode=disable"
$env:JWT_SECRET = "replace-with-a-long-random-secret"
```

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | Required at startup. PostgreSQL connection string. |
| `JWT_SECRET` | Required to issue and validate JWTs. Use a strong secret. |

The credentials above match the local Compose configuration. The API reads process environment variables directly; it does **not** load `.env` files automatically. `backend/.env` is ignored by Git.

### 3. Apply migrations

```powershell
migrate -path migrations -database "$env:DATABASE_URL" up
```

The three migrations create `vehicles`, create `users` (including the `pgcrypto` extension), and add the user role. Migrations are not applied automatically when the API starts. The database user must have permission to create the schema and extension.

### 4. Start the API

```powershell
go mod download
go run ./cmd/api
```

The API listens on `http://localhost:8080`. Port `8080` is hardcoded. Startup checks the database connection and fails if it cannot connect.

In another terminal:

```powershell
Invoke-RestMethod http://localhost:8080/health
```

The response contains `{"app":"GarageLog","status":"ok"}`. This endpoint is a liveness response; it does not perform a database query on each request.

## API endpoints

Protected requests require `Authorization: Bearer <token>`.

| Method | Path | Access | Success response |
| --- | --- | --- | --- |
| GET | `/health` | Public | `200`, application status |
| POST | `/auth/register` | Public | `201`, created user |
| POST | `/auth/login` | Public | `200`, token and user |
| GET | `/users/me` | Authenticated | `200`, current user |
| GET | `/admin/users` | Admin | `200`, user array |
| GET | `/vehicles` | Authenticated | `200`, vehicle array |
| GET | `/vehicles/:id` | Authenticated | `200`, vehicle |
| POST | `/vehicles` | Authenticated | `201`, created vehicle |
| PUT | `/vehicles/:id` | Authenticated | `200`, updated vehicle |
| DELETE | `/vehicles/:id` | Authenticated | `204`, no content |

Errors use a JSON object with an `error` field. Handlers return `400` for malformed request bodies or vehicle IDs, `401` for invalid credentials or tokens, `403` for non-admin access to the admin route, `404` for missing records, and `409` for duplicate registration emails. Database and other internal failures return `500`.

### Example session (PowerShell)

```powershell
$baseUrl = "http://localhost:8080"

$registration = @{
    first_name = "Alex"
    last_name = "Smith"
    email = "alex@example.com"
    password = "example-password-change-me"
} | ConvertTo-Json

Invoke-RestMethod "$baseUrl/auth/register" -Method Post `
    -ContentType "application/json" -Body $registration

$credentials = @{
    email = "alex@example.com"
    password = "example-password-change-me"
} | ConvertTo-Json

$login = Invoke-RestMethod "$baseUrl/auth/login" -Method Post `
    -ContentType "application/json" -Body $credentials
$headers = @{ Authorization = "Bearer $($login.token)" }

Invoke-RestMethod "$baseUrl/users/me" -Headers $headers

$vehicleBody = @{
    brand = "Toyota"
    model = "Corolla"
    year = 2020
    mileage = 65000
} | ConvertTo-Json

$vehicle = Invoke-RestMethod "$baseUrl/vehicles" -Method Post `
    -Headers $headers -ContentType "application/json" -Body $vehicleBody

Invoke-RestMethod "$baseUrl/vehicles" -Headers $headers
Invoke-RestMethod "$baseUrl/vehicles/$($vehicle.id)" -Headers $headers
```

Vehicle IDs are generated integers. User IDs are UUIDs. `PUT /vehicles/:id` writes all four vehicle fields (`brand`, `model`, `year`, `mileage`); it is not a partial update.

### Local admin account

There is no admin provisioning endpoint or seeded admin account. After registering a local account, promote it directly in the development database (run from `backend`):

```powershell
docker compose exec postgres psql -U garagelog -d garagelog -c "UPDATE users SET role = 'admin' WHERE email = 'alex@example.com';"
```

Log in again to obtain a token with the updated role. Authorization uses the role stored in the token, so changing a database role does not update an already-issued token.

## Repository layout

```text
backend/
  cmd/api/main.go          # API entry point and route registration
  internal/auth/          # JWT generation and authorization middleware
  internal/database/      # PostgreSQL connection setup
  internal/user/          # User models and handlers
  internal/vehicle/       # Vehicle model and CRUD handlers
  migrations/             # SQL schema migrations (up/down)
  docker-compose.yml      # Local PostgreSQL service
  go.mod
  go.sum
```

## Checks

Run from `backend`:

```powershell
go test ./...
go vet ./...
```

There are currently no automated test files. `go test ./...` checks that the packages compile, but does not establish end-to-end API or database correctness.

## Current limitations

- No frontend, maintenance history, service records, or expense tracking.
- No per-user vehicle ownership, list pagination, or filtering.
- No explicit required-field, email-format, password-strength, or vehicle-range validation in request models.
- No refresh tokens, logout/token revocation, password reset, or email verification.
- Registration returns an empty `role` string because its SQL statement does not return that column; the database still assigns `user`, and login/profile responses load it correctly.
- No automatic migrations, API container image, CI workflow, or deployment configuration in this repository.

This README describes the current implementation; the missing features above are not implemented functionality.
