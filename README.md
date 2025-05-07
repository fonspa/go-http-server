# Chirpy

Boot.dev Go HTTP server project.

Goals of this project:
- Understand what web servers are and how they power real-world web applications
- Build a production-style HTTP server in Go, without the use of a framework
- Use JSON, headers, and status codes to communicate with clients via a RESTful API
- Learn what makes Go a great language for building fast web servers
- Use type safe SQL to store and retrieve data from a Postgres database
- Implement a secure authentication/authorization system with well-tested cryptography libraries
- Build and understand webhooks and API keys
- Document the REST API with markdown

We'll be using PostgreSQL for the DB, [Goose](https://github.com/pressly/goose) for the migrations and [sqlc](https://sqlc.dev/) for generating Go code from SQL queries.

## Install

- Requires Go 1.22+
- Requires Goose
```shell
go install github.com/pressly/goose/v3/cmd/goose@latest
```
- Requires sqlc
```shell
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```
- Requires PostgreSQL 14+
```shell
sudo apt update
sudo apt install postgresql postgresql-contrib
```
  - Check that the installation worked:
```shell
psql --version
```
  - Set postgres password
```shell
sudo passwd postgres
```
  - Start Postgres service
```shell
sudo service postgresql start
```

### Creating Chirpy Database

- Enter psql shell:
```shell
$ sudo -u postgres psql
```
- Create `chirpy` database:
```sql
CREATE DATABASE chirpy;
```
- Connect to db:
```shell
\c chirpy
```
- Set user password
```sql
ALTER USER postgres WITH PASSWORD 'postgres';
```

Exiting psql shell with `exit`.

Your local database connection string should be `postgres://username:password@localhost:5432/chirpy`

You should be able to connect to your database with `psql` like:
```shell
$ psql postgres://username:password@localhost:5432/chirpy
```

Create now a `.env` file at the root of the repo, where you'll place the connection string:
```ini
DB_URL="postgres://username:password@localhost:5432/chirpy?sslmode=disable"
```

Make sure the `.env` file is in your `.gitignore`!

Install `godotenv`:
```shell
go get github.com/joho/godotenv
```
Then load the `.env` file from your code, in the `main`:
```Go
godotenv.Load()
```

## Running migrations with Goose

`cd` to `sql/schema` then, for the `Up` migrations:
```shell
goose postgres postgres://username:password@localhost:5432/chirpy up
```

Same thing with `down` for the `Down` migrations.

## SQLC

Create a Yaml configuration file `sqlc.yaml` for sqlc at the root of the repo:
```yaml
version: "2"
sql:
  - schema: "sql/schema"
    queries: "sql/queries"
    engine: "postgresql"
    gen:
      go:
        out: "internal/database"
```

sqlc will use `google/uuid` package to generate the UUIDs, so `go get` it:
```shell
go get github.com/google/uuid
```

We'll need a PostgreSQL driver too:
```shell
go get github.com/lib/pq
```

Write your queries in `sql/queries` folder.

Generate Go code from your queries with:
```shell
# From the root folder of the repo
sqlc generate
```
The Go code will be generated in the `internal/database` folder.
