#!/bin/bash

# Spanner Emulator用のインスタンスとデータベースを作成するスクリプト

set -e

PROJECT_ID="stunning-grin-480914-n1"
INSTANCE_ID="stunning-grin-480914-n1-instance"
DATABASE_ID="stunning-grin-480914-n1-db"

echo "==> Setting up Spanner Emulator..."

# Set emulator host
export SPANNER_EMULATOR_HOST="localhost:9010"

echo "==> Creating instance: $INSTANCE_ID"
gcloud spanner instances create $INSTANCE_ID \
  --config=emulator-config \
  --description="Visitas Development Instance" \
  --nodes=1 \
  --project=$PROJECT_ID || echo "Instance already exists"

echo "==> Creating database: $DATABASE_ID"
gcloud spanner databases create $DATABASE_ID \
  --instance=$INSTANCE_ID \
  --project=$PROJECT_ID || echo "Database already exists"

echo "==> Applying migrations..."

# Apply migrations in order
for migration in migrations/*.sql; do
  echo "Applying migration: $migration"
  gcloud spanner databases ddl update $DATABASE_ID \
    --instance=$INSTANCE_ID \
    --project=$PROJECT_ID \
    --ddl-file=$migration || echo "Migration may have already been applied: $migration"
done

echo "==> Spanner Emulator setup complete!"
echo ""
echo "Connection details:"
echo "  Project: $PROJECT_ID"
echo "  Instance: $INSTANCE_ID"
echo "  Database: $DATABASE_ID"
echo "  Emulator: $SPANNER_EMULATOR_HOST"
