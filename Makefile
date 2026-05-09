.PHONY: build run test dev clean docker-build docker-up docker-down docker-logs

# 自动检测可执行文件后缀
ifeq ($(OS),Windows_NT)
	EXT := .exe
else
	EXT :=
endif

build:
	go build -o bin/server$(EXT) ./cmd/server

run: build
	./bin/server$(EXT)

test:
	go test -v ./...

dev:
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
