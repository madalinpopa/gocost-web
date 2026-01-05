#!/usr/bin/env sh

set -e

# Set default values
DB_PATH=${DB_PATH:-/app/data/data.sqlite}
APP_ENV=${APP_ENV:-production}
APP_ADDR=${APP_ADDR:-0.0.0.0}
APP_PORT=${APP_PORT:-4000}

# Ensure data and uploads directories exist
mkdir -p "$(dirname "$DB_PATH")" /app/uploads

# Run migrations
./gocost migrate --dsn "$DB_PATH"

# Start the main application
exec ./main -addr "$APP_ADDR" -port "$APP_PORT" -dsn "$DB_PATH" -env "$APP_ENV"
