# Datastore code

Type-safe code from SQL is generated using [sqlc](https://github.com/kyleconroy/sqlc).

`sqlc` is configured through the file [sqlc.yalm](sqlc.yaml).

When there is a change in the schema or in the queries, the DB code should be re-generated:

```
make generate-sql-code
```

This regenerates the files `models.go`, `db.go`, `querier.go`, and files ending in `.sql.go`.

**These files should be committed.**

# Migrations

Migrations are done using [golang-migrate](https://github.com/golang-migrate/migrate).

When a new Datastore object is created using the NewDatastore method, the schema version is verified and
migrations are applied if needed.

# Change the DB schema

Once there was an initial release of Galadriel, changes in the DB schema should be added through new files
in the [migrations](postgres/migrations) folder and the queries in the [queries](postgres/queries) should be updated.
For SQLite, the files are in the [sqlite](sqlite) folder.