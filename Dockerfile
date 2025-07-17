# Stage 1: The Builder
FROM golang:1.24-alpine AS builder

# Set Timezone and install tzdata
ENV TZ=Asia/Jakarta
RUN apk add --no-cache tzdata

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server .

# ---

# Stage 2: The Final Image
FROM alpine:latest

# Set Timezone and install tzdata
ENV TZ=Asia/Jakarta
RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime && \
    echo ${TZ} > /etc/timezone

WORKDIR /app

# Copy compiled binary and assets
COPY --from=builder /app/server .
COPY --from=builder /app/config.json .
COPY --from=builder /app/certificate ./certificate

# Expose application port
EXPOSE 5000

# Run the application
CMD ["./server"]
