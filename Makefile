all: build test

build:
	go build -o main cmd/api/main.go

run:
	go run cmd/api/main.go

docker-run:
	docker compose up --build

docker-down:
	docker compose down

test:
	go test ./... -coverpkg=./... -count=1 -coverprofile coverage.out
	go tool cover -html coverage.out -o coverage.html

clean:
	rm -f main

.PHONY: all build run test clean docker-run docker-down
