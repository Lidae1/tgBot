#!/bin/bash

set -e

echo "=== Running migrations in Docker ==="

# Load environment variables
if [ -f .env.docker ]; then
    echo "Loading .env.docker file..."
    export $(cat .env.docker | grep -v '^#' | xargs)
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL environment variable is not set"
    exit 1
fi

echo "Running migrations..."
docker-compose exec crypto-bot ./main --migrate