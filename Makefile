.PHONY: swag build run test dev clean docker-build docker-up docker-down docker-logs

# 自动检测可执行文件后缀
ifeq ($(OS),Windows_NT)
	EXT := .exe
else
	EXT :=
endif

swag:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go

build: swag
	go build -o bin/server$(EXT) ./cmd/server

run: build
	./bin/server$(EXT)

test:
	go test -v ./...

dev: swag
	go run ./cmd/server -c config.yaml

clean:
	rm -rf bin/
	rm -f data/*.db data/*.db-shm data/*.db-wal

docker-build:
	docker build -t todo-app:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
