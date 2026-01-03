# Copilot Coding Agent Instructions

This document provides instructions for AI coding agents working on this repository.

## Project Overview

This is a **Golang/Gin** implementation of the [RealWorld](https://github.com/gothinkster/realworld) example application. It demonstrates a fully fledged fullstack application including CRUD operations, authentication, routing, pagination, and more.

## Technology Stack

- **Go**: 1.21+ required
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) v1.10+
- **ORM**: [GORM v2](https://gorm.io/) (gorm.io/gorm)
- **Database**: SQLite (for development/testing)
- **Authentication**: JWT using [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)
- **Validation**: [go-playground/validator/v10](https://github.com/go-playground/validator)

## Directory Structure

```
.
├── hello.go           # Main entry point
├── common/            # Shared utilities
│   ├── database.go    # Database connection manager
│   ├── utils.go       # Helper functions (JWT, validation, etc.)
│   └── unit_test.go   # Common package tests
├── users/             # User module
│   ├── models.go      # User data models & DB operations
│   ├── serializers.go # Response formatting
│   ├── routers.go     # Route handlers
│   ├── middlewares.go # Auth middleware
│   ├── validators.go  # Input validation
│   └── unit_test.go   # User package tests
├── articles/          # Articles module
│   ├── models.go      # Article data models & DB operations
│   ├── serializers.go # Response formatting
│   ├── routers.go     # Route handlers
│   ├── validators.go  # Input validation
│   └── unit_test.go   # Article package tests
└── scripts/           # Build/test scripts
```

## Development Commands

```bash
# Install dependencies
go mod download

# Build
go build ./...

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Format code
go fmt ./...

# Run linter
golangci-lint run

# Start the server
go run hello.go
```

## Code Style Guidelines

1. **Error Handling**: Always handle errors explicitly. Do not ignore errors with `_`.
2. **GORM v2 Patterns**:
   - Use `Preload()` instead of `Related()` for eager loading
   - Use `Association().Find()` for many-to-many relationships
   - Use `Updates()` instead of `Update()` for struct/map updates
   - Use pointers with `Delete()`: `Delete(&Model{})`
   - Count returns `int64`, handle overflow when converting to `uint`
3. **Validation Tags**: Use `required` instead of deprecated `exists` tag
4. **JWT**: Use `jwt.NewWithClaims()` for token creation

## Testing

- Tests are in `*_test.go` files alongside source code
- Use `common.TestDBInit()` and `common.TestDBFree()` for test database setup/teardown
- Run `go test ./...` before committing changes

## API Endpoints

The API follows the [RealWorld API Spec](https://realworld-docs.netlify.app/docs/specs/backend-specs/endpoints):

- `POST /api/users` - Register
- `POST /api/users/login` - Login
- `GET /api/user` - Get current user
- `PUT /api/user` - Update user
- `GET /api/profiles/:username` - Get profile
- `POST /api/profiles/:username/follow` - Follow user
- `DELETE /api/profiles/:username/follow` - Unfollow user
- `GET /api/articles` - List articles
- `GET /api/articles/feed` - Feed articles
- `GET /api/articles/:slug` - Get article
- `POST /api/articles` - Create article
- `PUT /api/articles/:slug` - Update article
- `DELETE /api/articles/:slug` - Delete article
- `POST /api/articles/:slug/comments` - Add comment
- `GET /api/articles/:slug/comments` - Get comments
- `DELETE /api/articles/:slug/comments/:id` - Delete comment
- `POST /api/articles/:slug/favorite` - Favorite article
- `DELETE /api/articles/:slug/favorite` - Unfavorite article
- `GET /api/tags` - Get tags

## Security Considerations

- JWT secret is defined in `common/utils.go` - do not expose in production
- Always validate user input using validator tags
- Use parameterized queries (GORM handles this automatically)
