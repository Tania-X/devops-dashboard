# ── 阶段 1：编译 ──
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/api

# ── 阶段 2：运行 ──
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /server /server

VOLUME ["/data"]
ENV DB_PATH=/data/devops.db

EXPOSE 8080
CMD ["/server"]
