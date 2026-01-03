# ![RealWorld Example App](logo.png)


[![CI](https://github.com/gothinkster/golang-gin-realworld-example-app/actions/workflows/ci.yml/badge.svg)](https://github.com/gothinkster/golang-gin-realworld-example-app/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/gothinkster/golang-gin-realworld-example-app/badge.svg?branch=main)](https://coveralls.io/github/gothinkster/golang-gin-realworld-example-app?branch=main)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/gothinkster/golang-gin-realworld-example-app/blob/main/LICENSE)
[![GoDoc](https://godoc.org/github.com/gothinkster/golang-gin-realworld-example-app?status.svg)](https://godoc.org/github.com/gothinkster/golang-gin-realworld-example-app)

> ### Golang/Gin codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.


This codebase was created to demonstrate a fully fledged fullstack application built with **Golang/Gin** including CRUD operations, authentication, routing, pagination, and more.

## Recent Updates (2026)

This project has been modernized with the following updates:
- **Go 1.21+**: Updated from Go 1.15 to require Go 1.21 or higher
- **GORM v2**: Migrated from deprecated jinzhu/gorm v1 to gorm.io/gorm v2
- **JWT v5**: Updated from deprecated dgrijalva/jwt-go to golang-jwt/jwt/v5 (fixes CVE-2020-26160)
- **Validator v10**: Updated validator tags and package to match gin v1.10.0
- **Latest Dependencies**: All dependencies updated to their 2025 production-stable versions
- **RealWorld API Spec Compliance (2026)**:
  - `GET /profiles/:username` now supports optional authentication (anonymous access allowed)
  - `POST /users/login` returns 401 Unauthorized on failure (previously 403)
  - `GET /articles/feed` registered as dedicated authenticated route
  - `DELETE /articles/:slug` and `DELETE /articles/:slug/comments/:id` return empty response body

## Dependencies (2025 Stable Versions)

| Package | Version | Release Date | Known Issues |
|---------|---------|--------------|--------------|
| [gin-gonic/gin](https://github.com/gin-gonic/gin) | v1.10.0 | 2024-05 | None; v1.11 has experimental HTTP/3 support |
| [gorm.io/gorm](https://gorm.io/) | v1.25.12 | 2024-08 | None; v1.30+ has breaking changes |
| [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) | v5.2.1 | 2024-06 | None; v5.3 only bumps Go version requirement |
| [go-playground/validator/v10](https://github.com/go-playground/validator) | v10.24.0 | 2024-12 | None; v10.30+ requires Go 1.24 |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | v0.32.0 | 2025-01 | None; keep updated for security fixes |
| [gorm.io/driver/sqlite](https://github.com/go-gorm/sqlite) | v1.5.7 | 2024-09 | None; requires cgo; use glebarez/sqlite for pure Go |
| [gosimple/slug](https://github.com/gosimple/slug) | v1.15.0 | 2024-12 | None |
| [stretchr/testify](https://github.com/stretchr/testify) | v1.10.0 | 2024-10 | None; v2 still in development |


# Directory structure
```
.
├── gorm.db
├── hello.go
├── common
│   ├── utils.go        //small tools function
│   └── database.go     //DB connect manager
├── users
|   ├── models.go       //data models define & DB operation
|   ├── serializers.go  //response computing & format
|   ├── routers.go      //business logic & router binding
|   ├── middlewares.go  //put the before & after logic of handle request
|   └── validators.go   //form/json checker
├── ...
...
```

# Getting started

## Install Golang

Make sure you have Go 1.21 or higher installed.

https://golang.org/doc/install

## Environment Config

Set-up the standard Go environment variables according to latest guidance (see https://golang.org/doc/install#install).


## Install Dependencies
From the project root, run:
```
go build ./...
go test ./...
go mod tidy
```

## Testing
From the project root, run:
```
go test ./...
```
or
```
go test ./... -cover
```
or
```
go test -v ./... -cover
```
depending on whether you want to see test coverage and how verbose the output you want.

## Todo
- More elegance config
- Test coverage (common & users 100%, article 0%)
- ProtoBuf support
- Code structure optimize (I think some place can use interface)
- Continuous integration (done)
