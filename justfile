DB_URL := 'postgres://postgres:admin@localhost:7765/chat?sslmode=disable&x-migrations-table=schema_migrations_auth'

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

migrate-up:
    migrate -path migrations -database "{{DB_URL}}" up

migrate-down:
    migrate -path migrations -database "{{DB_URL}}" down

migrate-force VERSION:
    migrate -path migrations -database "{{DB_URL}}" force {{VERSION}}

migrate-version:
    migrate -path migrations -database "{{DB_URL}}" version

migrate-create TITLE:
    migrate create --dir migrations --ext sql --seq {{TITLE}}

run:
    go run cmd/service-auth/main.go -config config/local.yaml