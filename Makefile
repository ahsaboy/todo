.PHONY: check-swag fix-swagger-docs swag build build-linux build-windows build-darwin run test dev clean docker-build docker-up docker-down docker-logs

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
APP    := server

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

# 编译当前平台
build: swag
ifeq ($(OS),Windows_NT)
	set "GOOS=$(GOOS)" && set "GOARCH=$(GOARCH)" && go build -o "bin\$(APP)-$(GOOS)-$(GOARCH)$(EXT)" ".\cmd\server"
else
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o "bin/$(APP)-$(GOOS)-$(GOARCH)$(EXT)" ./cmd/server
endif

# 交叉编译 Linux
build-linux: swag
ifeq ($(OS),Windows_NT)
	set "GOOS=linux" && set "GOARCH=amd64" && go build -o "bin\$(APP)-linux-amd64" ".\cmd\server"
else
	GOOS=linux GOARCH=amd64 go build -o "bin/$(APP)-linux-amd64" ./cmd/server
endif

# 交叉编译 Windows
build-windows: swag
ifeq ($(OS),Windows_NT)
	set "GOOS=windows" && set "GOARCH=amd64" && go build -o "bin\$(APP)-windows-amd64.exe" ".\cmd\server"
else
	GOOS=windows GOARCH=amd64 go build -o "bin/$(APP)-windows-amd64.exe" ./cmd/server
endif

# 交叉编译 macOS
build-darwin: swag
ifeq ($(OS),Windows_NT)
	set "GOOS=darwin" && set "GOARCH=arm64" && go build -o "bin\$(APP)-darwin-arm64" ".\cmd\server"
else
	GOOS=darwin GOARCH=arm64 go build -o "bin/$(APP)-darwin-arm64" ./cmd/server
endif

run: build
	.\bin\$(APP)-$(GOOS)-$(GOARCH)$(EXT)

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
