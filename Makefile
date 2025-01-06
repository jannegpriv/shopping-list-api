.PHONY: build run test test-docker clean docker-build docker-run docker-down docker-clean test-endpoints

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=shopping-list-api

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

run:
	$(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

test:
	$(GOTEST) -v ./...

test-docker: build
	docker-compose run --rm api-test

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v

# Development helpers
run-local:
	$(GOBUILD) -o $(BINARY_NAME)
	./$(BINARY_NAME)

# Test API endpoints
test-endpoints:
	@echo "Testing API endpoints..."
	@echo "\nCreating a new item:"
	curl -X POST http://localhost:8080/items \
		-H "Content-Type: application/json" \
		-d '{"name":"Milk","quantity":2,"price":3.99}'
	@echo "\n\nGetting all items:"
	curl http://localhost:8080/items
	@echo "\n"
	curl -X POST http://localhost:8080/items \
		-H "Content-Type: application/json" \
		-d '{"name":"Test Item","quantity":1,"price":9.99}'
	@echo "\n"
	curl http://localhost:8080/items
