# Multi-stage build

# Builder stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Runner stage
FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/main /

EXPOSE 8080

ENTRYPOINT ["/main"]