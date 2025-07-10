# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary from cmd/server
RUN CGO_ENABLED=0 go build -o main cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/main .

# Copy config if exists
COPY --from=builder /app/config.yaml ./config.yaml

EXPOSE 8080

CMD ["./main"]