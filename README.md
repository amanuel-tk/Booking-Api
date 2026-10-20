# Booking API

A movie ticket booking REST API written in Go. Admins manage movies, theaters, screens and showtimes. Users browse what's playing, pick seats, hold them for a few minutes, and confirm their booking.

The main engineering goal is **correctness under concurrency**: two people clicking the same seat at the same moment must never both get it.

**[Interactive API docs](https://<user>.github.io/<repo>/)** | [OpenAPI spec](docs/swagger.json)

## What it does

**For customers**
- Register and log in
- Browse movies, with search and pagination
- See showtimes for a movie on a given date
- See which seats are free, held or booked for a showtime
- Hold seats temporarily, then confirm the booking

**For admins**
- Add movies
- Add theaters and screens, where a screen is created from a grid size (for example 10 rows x 12 seats) and all its seats are generated automatically
- Schedule showtimes, where the end time is calculated from the movie's duration and overlapping shows on the same screen are rejected

## Highlights

- **No double booking.** Seat holds live in Redis with a TTL, and a Postgres unique constraint on `(showtime_id, seat_id)` is the final guarantee, even if a hold expires mid-request.
- **All-or-nothing holds.** If you ask for three seats and one is taken, none are held.
- **Overlap protection in the database.** Showtimes use a Postgres exclusion constraint, so two simultaneous requests can't both insert a clashing slot.
- **Transactional seat generation.** A screen and its seats are created in one transaction using `COPY`. A failure halfway leaves nothing behind.
- **Secure auth.** Bcrypt password hashing, short-lived JWT access tokens, role-based access (`user`, `admin`), and the same error for "unknown email" and "wrong password".
- **Clean layering.** Handler (HTTP) -> service (rules) -> repository (SQL), wired together in one place with dependency injection.

## Tech stack

| Area | Choice |
|---|---|
| Language | Go |
| Router | chi v5 |
| Database | PostgreSQL 16 with pgx v5 (no ORM) |
| Holds and cache | Redis 7 |
| Migrations | golang-migrate |
| Auth | JWT (HS256) and bcrypt |
| Docs | Swagger / OpenAPI |
| Local infra | Docker Compose |

## Project structure

```
cmd/api/main.go     entry point and wiring
migrations/         SQL migrations
internal/
  auth/             JWT, password hashing, auth middleware
  user/             register, login, current user
  movie/            movies
  theater/          theaters, screens, seats
  showtime/         showtimes
  booking/          bookings and seat availability
  hold/             Redis seat holds
  httpx/            JSON helpers and error format
  config/ database/ server/
```

## How booking works

1. The client asks for a showtime's seat map and sees which seats are free, held or booked.
2. The client holds the chosen seats. Redis sets all of them atomically or none, with a TTL of about 10 minutes.
3. The client confirms. The API checks that this user owns the hold, then inserts the booking and its seats in one Postgres transaction.
4. If another booking already took a seat, the unique constraint rejects the insert and the client gets a `409`.
5. The hold is released after a successful booking.

## Getting started

**Prerequisites:** Go, Docker with Docker Compose, and the [`migrate`](https://github.com/golang-migrate/migrate) CLI.

**1. Create a `.env` file**

```env
POSTGRES_USER=app
POSTGRES_PASSWORD=secret
POSTGRES_DB=auth

PORT=8080
DATABASE_URL=postgres://app:secret@localhost:5432/auth?sslmode=disable
REDIS_ADDR=localhost:16379
REDIS_PASSWORD=change-me
JWT_SECRET=replace-with-a-long-random-string
ACCESS_TOKEN_TTL=15m
```

**2. Start Postgres and Redis**

```bash
docker compose up -d
```

**3. Run migrations**

```bash
migrate -path migrations -database "postgres://app:secret@localhost:5432/auth?sslmode=disable" up
```

**4. Run the API**

```bash
go run ./cmd/api
```

Check it with `curl http://localhost:8080/health`, which should return `{"postgres":"ok","redis":"ok"}`.

**5. Make an admin.** New accounts get the `user` role. Register, then promote yourself and log in again, since roles are stored in the token:

```sql
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.email = 'you@example.com' AND r.name = 'admin';
```

## API documentation

With the server running, Swagger UI is at **http://localhost:8080/swagger/index.html**.

To call protected endpoints, log in with `POST /api/v1/auth/login`, copy the `access_token`, click **Authorize**, and enter `Bearer <token>`.

![Swagger UI](docs/images/swagger-ui.png)

Regenerate the docs after changing handlers or request structs:

```bash
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

All errors use one shape: `{"error": "message"}`.

## Testing

```bash
go test -race ./...
```

## Roadmap

- Refresh tokens and logout
- Concurrency tests (many goroutines booking one seat, run with `-race`)
- Background cleanup worker with graceful shutdown
- Rate limiting on login and booking
- Dockerfile and CI (`go vet`, `go test -race`)

## API Documentation

<!-- markdown-swagger -->
 Endpoint                   | Method | Auth? | Description                                                                                                                                                                                                                                                                      
 -------------------------- | ------ | ----- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 `/auth/login`              | POST   | No    | Validates credentials and returns a 15-minute JWT access token.                                                                                                                                                                                                                  
 `/auth/register`           | POST   | No    |                                                                                                                                                                                                                                                                                  
 `/me/bookings`             | GET    | Yes   | Retrieve up to 100 bookings belonging to the authenticated user, ordered by creation date descending. Includes both confirmed and cancelled bookings.                                                                                                                            
 `/me/bookings/{id}`        | GET    | Yes   | Retrieve a booking by UUID. Only bookings belonging to the authenticated user can be accessed.                                                                                                                                                                                   
 `/me/bookings/{id}/cancel` | POST   | Yes   | Cancel a confirmed booking belonging to the authenticated user. Cancellation is allowed only when the showtime starts at least 2 hours from the current time. Successfully cancelled bookings release their seats for future bookings.                                           
 `/movies`                  | GET    | No    | Retrieve active movies with optional title search and pagination. Results are ordered by release date descending. Default page is 1, default limit is 20, and maximum limit is 100.                                                                                              
 `/movies`                  | POST   | Yes   | Create a new movie. Title is required (max 200 characters), duration must be between 1 and 600 minutes, and release_date must use YYYY-MM-DD format. Requires admin access.                                                                                                      
 `/movies/{id}`             | GET    | No    | Retrieve an active movie by its UUID. Returns 404 if the movie does not exist or is inactive.                                                                                                                                                                                    
 `/movies/{id}/showtimes`   | GET    | No    | Retrieve upcoming showtimes for a movie within a date range. Dates use YYYY-MM-DD. The "date" parameter is an alternative to "from". If no dates are provided, today's date is used. The range is limited to 31 days.                                                            
 `/screens/{id}`            | GET    | No    | Retrieve a screen and its seats grouped by row. Each row contains its label and corresponding seats.                                                                                                                                                                             
 `/showtimes`               | POST   | Yes   | Schedule a movie on a screen. The start time must be in the future. The end time is calculated from the movie duration plus a 15-minute cleaning buffer. Overlapping showtimes on the same screen are rejected. Requires admin access.                                           
 `/showtimes/{id}`          | GET    | No    | Retrieve an active showtime by its UUID, including its movie, screen, start time, end time, and ticket price.                                                                                                                                                                    
 `/showtimes/{id}/bookings` | POST   | Yes   | Create a confirmed booking for 1 to 6 seats. The user must own valid Redis holds for all requested seats. The showtime must be in the future and all seats must belong to its screen. The total price is calculated by the server.                                               
 `/showtimes/{id}/holds`    | POST   | Yes   | Temporarily reserve 1 to 6 seats for the authenticated user. The showtime must exist and must not have started. All seats must belong to the showtime's screen and must not already be booked or held by another user. Holds expire automatically after the configured Redis TTL.
 `/showtimes/{id}/seats`    | GET    | No    | Retrieve all seats for a showtime grouped by row. Each seat has a status of free, held, or booked. Booked seats take precedence over temporary Redis holds.                                                                                                                      
 `/theaters`                | GET    | No    | Retrieve all active theaters ordered by creation date. Returns an empty array if no theaters exist.                                                                                                                                                                              
 `/theaters`                | POST   | Yes   | Create a new theater. The theater name and city must be provided and cannot exceed 100 characters. Requires admin access.                                                                                                                                                        
 `/theaters/{id}/screens`   | GET    | No    | Retrieve all active screens belonging to a theater. Returns an empty array if no active screens exist.                                                                                                                                                                           
 `/theaters/{id}/screens`   | POST   | Yes   | Create a screen in a theater and automatically generate its seats. Rows must be between 1 and 26, and seats per row must be between 1 and 50. Requires admin access.                                                                                                             
<!-- /markdown-swagger -->
