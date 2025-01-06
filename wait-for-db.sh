#!/bin/sh

set -e

host="${DB_HOST:-mysql}"
port="${DB_PORT:-3306}"
cmd="$@"

until nc -z -v -w30 $host $port
do
  echo "Waiting for database connection at ${host}:${port}..."
  sleep 5
done

echo "Database is up - executing command"
exec $cmd
