run:
	@./bin/goapps

run-dev:
	@go run cmd/api/main.go

build:
	@go build -o bin/goapps cmd/api/main.go