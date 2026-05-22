ARG APP_VERSION=dev

# 前端构建阶段
FROM node:20-alpine AS frontend-builder

ARG APP_VERSION
ENV VITE_APP_VERSION=${APP_VERSION}
ENV VITE_API_BASE_URL=/api/v1

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

ARG APP_VERSION
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${APP_VERSION}" -o /server ./cmd/server

# 运行阶段
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

LABEL org.opencontainers.image.source="https://github.com/ahsaboy/todo"

RUN mkdir -p /app/data

COPY --from=builder /server .

# 复制配置文件模板（实际配置通过 volume 挂载）
COPY config.example.yaml ./config.yaml

EXPOSE 8080

CMD ["./server"]
