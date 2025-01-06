FROM golang:1.21.5-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/wait-for-db.sh .

RUN chmod +x wait-for-db.sh

ENV ENV=production

EXPOSE 8080

CMD ["./wait-for-db.sh", "./main"]
