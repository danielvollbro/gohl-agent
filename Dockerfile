# Build Stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc libc-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o agent cmd/gohl/main.go

# Runtime Stage
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/agent .

CMD ["./agent", "scan", "--submit"]
