# Stage 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency definition files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build clean static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Minimal runtime environment (~15MB image size)
FROM alpine:latest
WORKDIR /app

# Copy compiled binary from builder stage
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]