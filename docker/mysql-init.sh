#!/bin/bash
set -e

echo "Running database migrations..."
for f in /migrations/*.up.sql; do
    echo "  Executing $(basename "$f")..."
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < "$f"
done
echo "Migrations complete."
