create-val-postgres:
    docker volume create postgres-data

postgres-run:
    docker run -d --name postgres-chat \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_DB=chat \
    -e POSTGRES_PASSWORD=admin \
    -p 7765:5432  \
    -v postgres-data:/var/lib/postgresql\
    -d postgres:18-bookworm

migrations-up:
    migrate -path migrations -database "postgres://postgres:admin@localhost:7765/chat?sslmode=disable" up

migrations-down:
    migrate -path migrations -database "postgres://postgres:admin@localhost:7765/chat?sslmode=disable" down

migrate-force VERSION:
    migrate -path migrations -database "postgres://postgres:admin@localhost:7765/chat?sslmode=disable" force {{VERSION}}

migrate-version:
    migrate -path migrations -database "postgres://postgres:admin@localhost:7765/chat?sslmode=disable" version

migrate-create TITLE:
    migrate create --dir migrations --ext sql --seq {{TITLE}}