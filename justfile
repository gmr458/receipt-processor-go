set shell := ["bash", "-cu"]
set dotenv-load

# --------------------------------------------------------------------
# Configuration
# --------------------------------------------------------------------

project := "webservice"
binary := "bin/" + project
cmd := "./cmd/webservice"
migrations := justfile_directory() / "postgres/migrations"

version := `git describe --always --dirty --tags --long`

migrate := "migrate"

# --------------------------------------------------------------------
# Development
# --------------------------------------------------------------------

# Run the web service.
run:
    go run {{ cmd }}

# Build the web service.
build:
    rm -rf bin
    mkdir -p bin
    go build \
        -ldflags "-s -w -X main.version={{ version }}" \
        -o {{ binary }} \
        {{ cmd }}

# Build and start the web service.
start: build
    ./{{ binary }}

format:
    gofmt -w -l .

# Run golangci-lint.
lint:
    golangci-lint run

# --------------------------------------------------------------------
# Database migrations
# --------------------------------------------------------------------

# Apply all pending migrations.
migrate-up:
    {{ migrate }} \
        -path "{{ migrations }}" \
        -database "$DSN" \
        up

# Roll back the last migration.
migrate-down:
    {{ migrate }} \
        -path "{{ migrations }}" \
        -database "$DSN" \
        down 1

# Force the database to a specific migration version.
migrate-force version:
    {{ migrate }} \
        -path "{{ migrations }}" \
        -database "$DSN" \
        force {{ version }}

# Show the current migration version.
migrate-version:
    {{ migrate }} \
        -path "{{ migrations }}" \
        -database "$DSN" \
        version
