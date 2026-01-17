#!/bin/bash

# Load environment variables from .env
export $(cat .env | xargs)

# Run migrations in order
echo "Running database migrations..."

echo "1. Creating users table..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -f schema/001_create_users.sql

echo "2. Creating sessions table..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -f schema/002_create_sessions.sql

echo "3. Altering quotes table..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -f schema/003_alter_quotes.sql

echo "4. Creating update triggers..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -f schema/004_update_triggers.sql

echo "5. Creating initial user and assigning quotes..."
envsubst < schema/005_initial_user.sql | PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT

echo "6. Adding foreign key constraint..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -f schema/006_add_foreign_key.sql

echo "7. Adding notes column to quotes..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -f schema/007_add_notes.sql

echo "Migrations complete!"
