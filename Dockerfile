# Stage 1: Build web
FROM node:24-alpine AS web-builder

WORKDIR /app/web

COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

# Stage 2: Build server
FROM golang:1.26.8-alpine3.24 AS builder
RUN apk update && apk upgrade --no-cache
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /app/web/build ./web/build

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o dist/url-shortener main.go

RUN mkdir -p /data && chown -R 65532:65532 /data

# Stage 3: Final Image
FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /

COPY --from=builder --chown=nonroot:nonroot /data /data
COPY --from=builder /app/dist/url-shortener /url-shortener

VOLUME ["/data"]

ENV DB_PATH=/data/shortener.db
ENV PORT=3011

EXPOSE 3011

USER nonroot:nonroot

ENTRYPOINT ["/url-shortener"]