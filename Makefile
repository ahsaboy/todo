.PHONY: frontend-install frontend-build frontend-clean check-swag fix-swagger-docs swag build build-linux build-windows build-darwin run test dev clean docker-build docker-up docker-down docker-logs

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
APP    := todo
LDFLAGS := -s -w
BUILDFLAGS := -trimpath -ldflags="$(LDFLAGS)"

ifeq ($(OS),Windows_NT)
	SHELL := cmd.exe
.SHELLFLAGS := /C
	EXT := .exe
else
	EXT :=
endif

# 检测 UPX 是否可用
ifeq ($(OS),Windows_NT)
	HAS_UPX := $(strip $(shell cmd /C "where upx >nul 2>nul && echo 1 || echo 0"))
else
	HAS_UPX := $(strip $(shell command -v upx >/dev/null 2>&1 && echo 1 || echo 0))
endif

# UPX 压缩（如果系统安装了 UPX 则自动压缩）
ifeq ($(OS),Windows_NT)
define upx_compress
	@if "$(HAS_UPX)"=="1" (echo UPX found, compressing... && upx --best --lzma "$(1)" 2>nul) else (echo UPX not found, skipping compression)
endef
else
define upx_compress
	@if [ "$(HAS_UPX)" = "1" ]; then \
		echo "UPX found, compressing..."; \
		upx --best --lzma $(1) 2>/dev/null || true; \
	else \
		echo "UPX not found, skipping compression"; \
	fi
endef
endif

# 前端依赖安装
frontend-install:
ifeq ($(OS),Windows_NT)
	cd /d frontend && npm ci
else
	cd frontend && npm ci
endif

# 前端构建
frontend-build: frontend-install
ifeq ($(OS),Windows_NT)
	cd /d frontend && npm run build
	if exist web\dist rmdir /s /q web\dist
	xcopy /E /I /Y frontend\dist web\dist
else
	cd frontend && npm run build && cd .. && rm -rf web/dist && cp -r frontend/dist web/dist
endif

# 前端清理
frontend-clean:
ifeq ($(OS),Windows_NT)
	if exist frontend\dist rmdir /s /q frontend\dist
	if exist web\dist rmdir /s /q web\dist
else
	rm -rf frontend/dist web/dist
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
build: frontend-build swag
ifeq ($(OS),Windows_NT)
	set "GOOS=$(GOOS)" && set "GOARCH=$(GOARCH)" && go build $(BUILDFLAGS) -o "bin\$(APP)-$(GOOS)-$(GOARCH)$(EXT)" ".\cmd\server"
else
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILDFLAGS) -o "bin/$(APP)-$(GOOS)-$(GOARCH)$(EXT)" ./cmd/server
endif
	$(call upx_compress,bin/$(APP)-$(GOOS)-$(GOARCH)$(EXT))

# 交叉编译 Linux
build-linux: frontend-build swag
ifeq ($(OS),Windows_NT)
	set "GOOS=linux" && set "GOARCH=amd64" && go build $(BUILDFLAGS) -o "bin\$(APP)-linux-amd64" ".\cmd\server"
else
	GOOS=linux GOARCH=amd64 go build $(BUILDFLAGS) -o "bin/$(APP)-linux-amd64" ./cmd/server
endif
	$(call upx_compress,bin/$(APP)-linux-amd64)

# 交叉编译 Windows
build-windows: frontend-build swag
ifeq ($(OS),Windows_NT)
	set "GOOS=windows" && set "GOARCH=amd64" && go build $(BUILDFLAGS) -o "bin\$(APP)-windows-amd64.exe" ".\cmd\server"
else
	GOOS=windows GOARCH=amd64 go build $(BUILDFLAGS) -o "bin/$(APP)-windows-amd64.exe" ./cmd/server
endif
	$(call upx_compress,bin/$(APP)-windows-amd64.exe)

# 交叉编译 macOS
build-darwin: frontend-build swag
ifeq ($(OS),Windows_NT)
	set "GOOS=darwin" && set "GOARCH=arm64" && go build $(BUILDFLAGS) -o "bin\$(APP)-darwin-arm64" ".\cmd\server"
else
	GOOS=darwin GOARCH=arm64 go build $(BUILDFLAGS) -o "bin/$(APP)-darwin-arm64" ./cmd/server
endif
	$(call upx_compress,bin/$(APP)-darwin-arm64)

run: build
	.\bin\$(APP)-$(GOOS)-$(GOARCH)$(EXT)

test:
	go test -v ./...

dev: frontend-build swag
	go run ./cmd/server -c config.yaml

clean:
	if exist bin rmdir /s /q bin
	if exist data\*.db del /q data\*.db
	if exist data\*.db-shm del /q data\*.db-shm
	if exist data\*.db-wal del /q data\*.db-wal
	if exist frontend\dist rmdir /s /q frontend\dist
	if exist web\dist rmdir /s /q web\dist

docker-build:
	docker build -t todo-app:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
