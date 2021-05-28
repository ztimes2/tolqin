#!/bin/bash

source .env

MIGRATIONS_SOURCE=file://migrations
DB_URL=postgres://$DB_USERNAME:$DB_USERNAME@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE

ACTION=$1
if [ $ACTION != 'up' ] && [ $ACTION != 'down' ]; then
    echo "invalid or missing argument: must be 'up' or 'down'"
    exit 1
fi

migrate --source $MIGRATIONS_SOURCE --database $DB_URL $ACTION
