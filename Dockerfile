# Use the official Golang image as the base for building
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk update && apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project into the image
COPY . .

# Build the client application
RUN go build -o client-app ./client

# Build the server application
RUN go build -o server-app ./server

# Use a minimal image for the runtime
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy built binaries from the builder stage
COPY --from=builder /app/client-app .
COPY --from=builder /app/server-app .

# Expose ports
EXPOSE 8080 9090

# Default command (can be overridden)
CMD ["./server-app"]
