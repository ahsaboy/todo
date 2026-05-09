# 前端构建阶段
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci

COPY frontend/ .
RUN npm run build

# Go 构建阶段
FROM golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 从前端构建阶段复制 dist 目录
COPY --from=frontend-builder /app/dist ./web/dist

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal && \
    sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# 运行阶段
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

RUN mkdir -p /app/data

COPY --from=builder /server .
COPY config.yaml .

EXPOSE 8080

CMD ["./server"]
