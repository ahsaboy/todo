ARG APP_VERSION=dev
ARG API_BASE_URL=/api/v1

FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .

ARG APP_VERSION
ARG API_BASE_URL
ENV VITE_APP_VERSION=${APP_VERSION}
ENV VITE_API_BASE_URL=${API_BASE_URL}
RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1/ || exit 1
