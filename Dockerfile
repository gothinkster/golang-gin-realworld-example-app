#build stage
FROM golang:latest AS builder
RUN apt-get install git
WORKDIR /go/src
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin -v ./... && \
    mv /go/bin/golang-gin-realworld-example-app /go/bin/app && \
    ls -l /go/bin

#final stage
FROM ubuntu:latest AS target
RUN apt update && apt -y upgrade
COPY --from=builder /go/bin/ /
ENTRYPOINT /app
LABEL Name=readlworldapp Version=0.0.1
EXPOSE 8080
