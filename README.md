# Nip - Go Microblogging Engine Similar to Twitter

Nip is a high-performance microblogging backend built with Go, demonstrating idiomatic project structure, dependency injection, and polyglot persistence (MySQL + MongoDB).

## Tech Stack

- **Core**: Go 1.22+
- **Primary Database (Users/Follows)**: MySQL
- **Messaging/Timeline Storage**: MongoDB
- **Infrastructure**: Docker & Docker Compose
- **Migrations**: `migrate` tool

## Project Structure

The project follows the "Domain-Driven Root" pattern:

- `/`: Core domain entities (`user.go`, `tweet.go`, `follow.go`) and service interfaces (`service.go`).
- `internal/`: Private implementations for repositories and business logic.
- `cmd/nip/`: Main orchestration and application entry point.
- `migrations/`: Database schema definitions.

## Getting Started

### 1. Start Infrastructure
```bash
docker-compose up -d
```

### 2. Run Database Migrations
Ensure the `migrate` tool is installed, then run:
```bash
migrate -path ./migrations -database "mysql://root:root@tcp(127.0.0.1:3306)/nip" up
```

### 3. Build and Run
```bash
go build ./cmd/nip
./nip
```

## Testing

The cmd/nip/scenario_test.go file contains the scenario that was in the assignment.
Follow the following instructions to run the test:

1. Start Infrastructure
2. Run Database Migrations
3. Build and Run
4. Run the test

### Clear Test Cache
If you want to force a fresh test run without using cached results:
```bash
go clean -testcache
```

### Run All Tests
```bash
go test -v ./...
```

## Development & API Documentation

Nip uses standard Go doc comments for all public structures and interfaces. You can view the documentation locally using:
```bash
# View root domain documentation
go doc -all

# View specific package documentation (e.g., services)
go doc -all ./internal/service
```

### Core Interfaces
- **AuthService**: Handles registration and OTP-based login.
- **TweetService**: Manages tweet creation, hashtag extraction, and mentions.
- **TimelineService**: Caches the timeline of users.
- **UserRepository/TweetRepository**: Abstract persistence layers.