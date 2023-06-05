# Datastore

The Galadriel project supports two datastore engines: SQLite and Postgres. The Datastore handles all database
interactions. This document outlines the procedures for generating type-safe code, handling migrations, and changing the
database schema.

## SQL Code Generation

We use [sqlc](https://github.com/kyleconroy/sqlc) to generate type-safe Go code for SQL queries. Sqlc configurations are
defined in the [sqlc.yaml](sqlc.yaml) file.

Whenever there are changes to the schema or queries, run the following command to regenerate the Go code:

```shell
make generate-sql-code
```

This command regenerates the `models.gen.go`, `db.gen.go`, `querier.gen.go`, and the `.sql.go` files.

**Note:** Remember to commit these regenerated files.

## Database Migrations

Database migrations are managed with [golang-migrate](https://github.com/golang-migrate/migrate).

During the creation of a new Datastore object using the `NewDatastore` method, the schema version is verified, and any
necessary migrations are applied.

## Changing the Database Schema

Following the initial release of Galadriel, any changes to the DB schema must be handled through new files in
the [postgres migrations](postgres/migrations) and [sqlite3 migrations](sqlite/migrations) folders. Also, the queries in
the [postgres queries](postgres/queries) and [sqlite queries](sqlite/queries) should be updated accordingly.

To reflect the current schema version supported by Galadriel and to ensure automatic migration, remember to increment
the `currentDBVersion` constant.
