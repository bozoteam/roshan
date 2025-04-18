#!/usr/bin/env fish

function cleanup
    echo "Cleaning up..."
    docker compose down
    # Kill any running go processes started by this script
    pkill -P (echo %self)
    exit 0
end

trap cleanup SIGINT SIGTERM

docker compose up -d; or cleanup

while true
    echo "Waiting for PostgreSQL to be ready..."
    pg_isready -U postgres -h 127.0.0.1
    if test $status -eq 0
        echo "PostgreSQL is ready!"
        break
    end
    sleep 1
end

./migrate.fish; or cleanup
./apply.fish; or cleanup

go run ./cmd/server/main.go; or cleanup
