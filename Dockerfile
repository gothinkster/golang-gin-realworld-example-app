#build stage
FROM golang:latest AS builder
RUN apt-get install git
WORKDIR /go/src
COPY . .
RUN go get -d -v ./...
RUN go build -race -o ./ -v ./... && \
    ls -lh ./golang*

#final stage
FROM ubuntu:latest AS target
WORKDIR /app
RUN apt update && apt -y upgrade
COPY --from=builder /go/src/golang-gin-realworld-example-app .
LABEL Name=readlworldapp Version=0.0.1
EXPOSE 8080
CMD ["/app/golang-gin-realworld-example-app"]
