#!/bin/bash

echo "Setting up RateX project..."

# Initialize go module if not exists
if [ ! -f go.mod ]; then
    go mod init github.com/afriwondimu/RateX
fi

# Install dependencies
echo "Installing dependencies..."
go get github.com/gin-gonic/gin@v1.10.0
go get github.com/go-redis/redis/v8@v8.11.5
go get github.com/google/uuid@v1.6.0
go get github.com/stretchr/testify@v1.9.0

# Tidy modules
echo "Tidying modules..."
go mod tidy

# Create necessary directories
echo "Creating directory structure..."
mkdir -p limiter/{algorithms,storage,middleware}
mkdir -p examples/{basic,redis,custom}
mkdir -p test
mkdir -p cmd/ratex

echo "Setup complete!"
echo ""
echo "To run tests:"
echo "  make test-short        # Run tests without Redis"
echo "  make test-integration   # Run tests with Redis (requires Redis running)"
echo ""
echo "To run examples:"
echo "  make run-example-basic"
echo "  make run-example-redis (requires Redis)"