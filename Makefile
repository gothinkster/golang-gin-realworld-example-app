.PHONY = build clean build-container clean-container run
.DEFAULT_GOAL = build
DOCKER = vctl
APPNAME = realworldapp

build: clean
	go get -d -v ./...
	go build -v -race -o ./build ./...
	go test ./...

clean:
	go clean

build-container:
	 $(DOCKER) build -t $(APPNAME):latest .

clean-container:
	$(DOCKER) rm -f rw
	$(DOCKER) rmi $(APPNAME)

run:
	$(DOCKER) run -d -p 8080:8080/tcp --name rw $(APPNAME):latest