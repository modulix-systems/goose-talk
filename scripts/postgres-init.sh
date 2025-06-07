#!/usr/bin/bash
set -e

test_user_username="test_user"
main_user_username="root"

databases=("auth") # will extend in future

docker compose exec postgres psql -U $main_user_username -c "CREATE ROLE test_user WITH LOGIN PASSWORD 'test';" || true

# create both main and test databases for every db name in the list
for dbname in "${databases[@]}"; do
  docker compose exec postgres psql -U $main_user_username -c "CREATE DATABASE $dbname WITH owner=$main_user_username" \
  -c "CREATE DATABASE ${dbname}_test WITH owner=$test_user_username" || true
  docker compose exec postgres psql -U $main_user_username -d "${dbname}" -c "CREATE EXTENSION IF NOT EXISTS CITEXT"
  docker compose exec postgres psql -U $main_user_username -d "${dbname}_test" -c "CREATE EXTENSION IF NOT EXISTS CITEXT"
done

echo "Postgres cluster succesfully initialized"
