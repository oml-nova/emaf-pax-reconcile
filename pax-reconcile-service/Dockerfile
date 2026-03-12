# ---- Builder stage ----
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Download dependencies separately for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build from cmd/ directory, statically linked for Alpine
WORKDIR /app/cmd
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# ---- Runtime stage ----
FROM alpine:3.19

# Create non-root user (UID 5555)
RUN addgroup -g 5555 plateron && adduser -D -u 5555 -G plateron plateron

WORKDIR /app

# Copy binary
COPY --from=builder /app/main .

EXPOSE 8080

# Set ownership then drop privileges
RUN chown -R plateron:plateron /app
USER plateron

CMD ["./main"]
