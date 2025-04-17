#!/usr/bin/env fish

docker compose up -d

while true
    echo "Waiting for PostgreSQL to be ready..."
    pg_isready -U postgres -h 127.0.0.1
    if test $status -eq 0
        echo "PostgreSQL is ready!"
        break
    end
    sleep 1
end

./migrate.fish
./apply.fish
