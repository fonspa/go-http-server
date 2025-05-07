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

We'll be using PostgreSQL for the DB, [Goose](https://github.com/pressly/goose) for the migrations.

## Install

- Requires Go 1.22+
- Requires Goose
```shell
go install github.com/pressly/goose/v3/cmd/goose@latest
```
- Requires PostgreSQL 14+
```shell
$ sudo apt update
$ sudo apt install postgresql postgresql-contrib
```
  - Check that the installation worked:
```shell
$ psql --version
```
  - Set postgres password
```shell
$ sudo passwd postgres
```
  - Start Postgres service
```shell
$ sudo service postgresql start
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

### Running migrations with Goose

`cd` to `sql/schema` then, for the `Up` migrations:
```shell
goose postgres postgres://username:password@localhost:5432/chirpy up
```

Same thing with `down` for the `Down` migrations.
