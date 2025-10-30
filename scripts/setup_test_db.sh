#!/bin/bash
# Setup test database with sample data for metrics examples.
# Safe to run multiple times - creates fresh database each time.

set -e

# Define database path (matches default from environment)
DB_PATH="${DB_PATH:-./data.db}"

# Remove existing database to start fresh
if [ -f "$DB_PATH" ]; then
    rm "$DB_PATH"
    echo "Removed existing database: $DB_PATH"
fi

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Run SQL file to create schema and insert sample data
sqlite3 "$DB_PATH" < "$SCRIPT_DIR/setup_test_db.sql"

echo "Test database created: $DB_PATH"
