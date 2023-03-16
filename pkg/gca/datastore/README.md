# Datastore code 

Type-safe code from SQL is generated using [sqlc](https://github.com/kyleconroy/sqlc).

`sqlc` is configured through the file [sqlc.yalm](sqlc.yaml).

When there is a change in the schema or in the queries, the DB code should be re-generated:

```
go get github.com/kyleconroy/sqlc/cmd/sqlc
go install github.com/kyleconroy/sqlc/cmd/sqlc
```

```
sqlc generate
```

This regenerates the files `models.go`, `db.go`, `querier.go`, and files ending in `.sql.go`.

**These files should be committed.**

# Migrations

Migrations are done using [golang-migrate](https://github.com/golang-migrate/migrate).

When a new Datastore object is created using the NewDatastore method, the schema version is verified and 
migrations are applied if needed. 

Schema validation and migrations use schema.go from the common/datastore module. 

# Change the DB schema

Once there was an initial release of Galadriel, changes in the DB schema should be added through new files
in the [migrations](migrations) folder and the queries in the [queries](queries) should be updated. The database version and schema is tracked within each local datastore. When a DDL type of change is needed, the version must be incremented and the files in the migration folder must be preappended with that new version number.