.PHONY: help frontend-install frontend-build frontend-build-standalone frontend-clean check-swag swag build build-linux build-windows build-darwin run test dev clean docker-build docker-up docker-down docker-logs

GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
APP    := todo
IMAGE  ?= ghcr.io/ahsaboy/todo:latest
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
LDFLAGS := -s -w
BUILDFLAGS := -trimpath -ldflags="$(LDFLAGS) -X main.version=$(GIT_TAG)"

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

help:
	@echo "TODO 任务管理系统 Makefile 命令"
	@echo ""
	@echo "常用:"
	@echo "  make help                         显示此帮助"
	@echo "  make dev                          本地运行，先构建前端并生成 Swagger"
	@echo "  make test                         运行 Go 测试"
	@echo "  make clean                        清理构建产物和本地数据库"
	@echo ""
	@echo "单体构建（后端嵌入前端静态资源）:"
	@echo "  make build                        编译当前平台"
	@echo "  make build-linux                  编译 Linux amd64"
	@echo "  make build-windows                编译 Windows amd64"
	@echo "  make build-darwin                 编译 macOS arm64"
	@echo ""
	@echo "独立前端静态资源:"
	@echo "  make frontend-build-standalone API_BASE_URL=https://api.example.com/api/v1"
	@echo "                                    仅构建前端 dist，不复制到 web/dist"
	@echo ""
	@echo "前端与文档:"
	@echo "  make frontend-install             安装前端依赖"
	@echo "  make frontend-build               构建前端并复制到 web/dist"
	@echo "  make frontend-clean               清理前端构建产物"
	@echo "  make swag                         生成 Swagger 文档"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build                 构建单体 Docker 镜像（默认标签: $(IMAGE)）"
	@echo "  make docker-up                    启动 docker compose"
	@echo "  make docker-down                  停止 docker compose"
	@echo "  make docker-logs                  查看 docker compose 日志"

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
	cd /d frontend && set "VITE_APP_VERSION=$(GIT_TAG)" && npm run build
	if exist web\dist rmdir /s /q web\dist
	xcopy /E /I /Y frontend\dist web\dist
else
	cd frontend && VITE_APP_VERSION=$(GIT_TAG) npm run build && cd .. && rm -rf web/dist && cp -r frontend/dist web/dist
endif

# 前端独立构建（不复制到 web/dist）
frontend-build-standalone: frontend-install
ifeq ($(OS),Windows_NT)
	cd /d frontend && set "VITE_APP_VERSION=$(GIT_TAG)" && set "VITE_API_BASE_URL=$(API_BASE_URL)" && npm run build
else
	cd frontend && VITE_APP_VERSION=$(GIT_TAG) VITE_API_BASE_URL="$(API_BASE_URL)" npm run build
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

swag: check-swag
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

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
ifeq ($(OS),Windows_NT)
	.\bin\$(APP)-$(GOOS)-$(GOARCH)$(EXT)
else
	./bin/$(APP)-$(GOOS)-$(GOARCH)$(EXT)
endif

test:
	go test -v ./...

dev: frontend-build swag
	go run ./cmd/server -c config.yaml

clean:
ifeq ($(OS),Windows_NT)
	if exist bin rmdir /s /q bin
	if exist data\*.db del /q data\*.db
	if exist data\*.db-shm del /q data\*.db-shm
	if exist data\*.db-wal del /q data\*.db-wal
	if exist frontend\dist rmdir /s /q frontend\dist
	if exist web\dist rmdir /s /q web\dist
else
	rm -rf bin
	rm -f data/*.db data/*.db-shm data/*.db-wal
	rm -rf frontend/dist web/dist
endif

docker-build:
	docker build --build-arg APP_VERSION=$(GIT_TAG) -t $(IMAGE) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
