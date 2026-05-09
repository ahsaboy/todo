.PHONY: build run test clean docker-build docker-up docker-down

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -f data/*.db data/*.db-shm data/*.db-wal

docker-build:
	docker build -t todo-app:latest .

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f
