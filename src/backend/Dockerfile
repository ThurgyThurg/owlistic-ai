# Use the official Golang image as the base image
FROM golang:1.23.4-alpine AS builder

# Set ARG for platform targeting (but don't force platform)
ARG TARGETARCH=amd64

# Install base compiler libraries
RUN apk add --no-cache \
    build-base \
    gcc \
    musl-dev

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first for better layer caching
COPY ./go.mod ./go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GO111MODULE=on \
    go build -v \
    -o /app/owlistic ./cmd/main.go

# Use a minimal Alpine image for the final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/owlistic ./

# Make sure the binary is executable
RUN chmod +x /app/owlistic

# Expose the application port
EXPOSE 8080

# Set the entrypoint to the binary
ENTRYPOINT ["/app/owlistic"]
