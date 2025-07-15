# Stage 1: The Builder
# We use a specific Go version to build our application.
FROM golang:1.24-alpine AS builder
ENV TZ=Asia/Jakarta
# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application.
# CGO_ENABLED=0 is important for creating a static binary that can run in a minimal image like alpine.
# -o /app/server creates the executable named 'server' in the /app directory.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server .

# ---

# Stage 2: The Final Image
# We start from a minimal Alpine Linux image. It's much smaller than the Go image.
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy ONLY the compiled binary from the 'builder' stage.
# This is the magic of multi-stage builds! We're not copying any source code or build tools.
COPY --from=builder /app/server .

# We also copy the default config and certificates. The config can be overwritten by a volume mount.
COPY --from=builder /app/config.json .
COPY --from=builder /app/certificate ./certificate

# Expose the port your Go application listens on (e.g., 5000).
EXPOSE 5000

# The command to run when the container starts.
# This executes the compiled binary.
CMD ["./server"]
