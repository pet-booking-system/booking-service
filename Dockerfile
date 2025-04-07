# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

# Stage 2: Minimal runtime
FROM alpine:latest
WORKDIR /root/

RUN apk add --no-cache postgresql-client

COPY --from=builder /app/app /app

RUN chmod +x /app

EXPOSE 50052
CMD ["/app"]
