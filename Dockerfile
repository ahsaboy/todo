FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal && \
    sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

RUN mkdir -p /app/data

COPY --from=builder /server .
COPY config.yaml .

EXPOSE 8080

CMD ["./server"]
