#!/bin/bash
set -e

# Set environment variables for Linux build
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# Build the binary
echo "Building Go binary for Linux..."
go build -o bootstrap main.go

# Create deployment package
echo "Creating deployment package..."
zip function.zip bootstrap

# Clean up binary
echo "Cleaning up..."
rm bootstrap

echo "Done! Deployment package created as function.zip"