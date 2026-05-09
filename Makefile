.PHONY: check-swag fix-swagger-docs swag build run test dev clean docker-build docker-up docker-down docker-logs

ifeq ($(OS),Windows_NT)
	SHELL := cmd.exe
	.SHELLFLAGS := /C
	EXT := .exe
else
	EXT :=
endif

check-swag:
ifeq ($(OS),Windows_NT)
	where swag >nul 2>nul || (echo swag is not installed. Install it with: go install github.com/swaggo/swag/cmd/swag@latest && exit /b 1)
	swag --version
else
	command -v swag >/dev/null 2>&1 || (echo "swag is not installed. Install it with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	swag --version
endif

fix-swagger-docs:
	go run ./scripts/fix_swagger_docs.go

swag: check-swag
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	$(MAKE) fix-swagger-docs

build: swag
	go build -o bin/server$(EXT) ./cmd/server

run: build
	.\bin\server$(EXT)

test:
	go test -v ./...

dev: swag
	go run ./cmd/server -c config.yaml

clean:
	if exist bin rmdir /s /q bin
	if exist data\*.db del /q data\*.db
	if exist data\*.db-shm del /q data\*.db-shm
	if exist data\*.db-wal del /q data\*.db-wal

docker-build:
	docker build -t todo-app:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
