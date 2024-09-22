# First stage: build the Go application
FROM golang:1.23.1 as builder

WORKDIR /go/src

# Copy your source code
COPY . .

# Build your application
RUN CGO_ENABLED=0 go build -mod=vendor -o app

# Use a second stage to prepare the runtime container
FROM debian:buster-slim as runtime

# Install CA certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built application from the builder stage
COPY --from=builder /go/src/app /app

# Non-root user
USER 1001

# Set the working directory
WORKDIR /

# Command to run
CMD ["/app"]