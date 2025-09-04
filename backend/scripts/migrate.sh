#!/bin/bash

# Database Migration Scripts for API Monitor
# Usage: ./scripts/migrate.sh [command]

case "$1" in
  fresh)
    echo "🔄 Running fresh migration (dropping all tables)..."
    go run cmd/migrate/main.go fresh
    ;;
  run)
    echo "🔄 Running pending migrations..."
    go run cmd/migrate/main.go run
    ;;
  *)
    echo "Usage: $0 {fresh|run}"
    echo ""
    echo "Commands:"
    echo "  fresh   Drop all tables and run all migrations (like Laravel's migrate:fresh)"
    echo "  run     Run pending migrations only"
    echo ""
    echo "Examples:"
    echo "  $0 fresh    # Fresh migration"
    echo "  $0 run      # Run migrations"
    exit 1
    ;;
esac
