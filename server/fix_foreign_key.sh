#!/bin/bash

# Load environment variables from .env
export $(cat .env | xargs)

echo "Dropping old foreign key constraint..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -c "ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_user_id_fkey;"

echo "Running 006 migration to add new foreign key..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOSTNAME -U $DB_USERNAME -d $DB_NAME -p $DB_PORT -f schema/006_add_foreign_key.sql

echo "Done!"
