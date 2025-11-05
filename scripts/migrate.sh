#!/bin/bash

set -e

echo "=== Database Migration Script ==="

# Load environment variables
if [ -f .env ]; then
    echo "Loading .env file..."
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable is not set"
    echo "Please set DATABASE_URL or create .env file"
    exit 1
fi

echo "Database URL: $(echo $DATABASE_URL | sed 's/\/\/.*@/\/\/***@/')" # Mask password

# Check if goose is installed
if ! command -v goose &> /dev/null; then
    echo "Error: goose is not installed. Install with: go install github.com/pressly/goose/v3/cmd/goose@latest"
    exit 1
fi

# Check if migrations directory exists
if [ ! -d "migrations" ]; then
    echo "Error: migrations directory not found"
    exit 1
fi

echo "Running database migrations..."
goose -dir migrations postgres "$DATABASE_URL" up

echo "✅ Migrations completed successfully!"

# Show migration status
echo ""
echo "Current migration status:"
goose -dir migrations postgres "$DATABASE_URL" status