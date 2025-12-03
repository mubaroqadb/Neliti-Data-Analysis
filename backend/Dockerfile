# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (using main-cloudrun.go instead of main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main-cloudrun.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy all necessary packages (config, controller, helper, model, route)
COPY --from=builder /app/config ./config
COPY --from=builder /app/controller ./controller
COPY --from=builder /app/helper ./helper
COPY --from=builder /app/model ./model
COPY --from=builder /app/route ./route

# Expose port (Cloud Run will set PORT env var)
EXPOSE 8080

# Run the application
CMD ["./main"]