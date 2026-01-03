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
- **Latest Dependencies**: All dependencies updated to their latest stable versions
- **RealWorld API Spec Compliance (2026)**:
  - `GET /profiles/:username` now supports optional authentication (anonymous access allowed)
  - `POST /users/login` returns 401 Unauthorized on failure (previously 403)
  - `GET /articles/feed` registered as dedicated authenticated route
  - `DELETE /articles/:slug` and `DELETE /articles/:slug/comments/:id` return empty response body


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
