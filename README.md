# sqlc: A SQL Compiler [Modified]

sqlc generates **fully type-safe idiomatic Go code** from SQL schema. Here's how it
works:

1. You write SQL schema
1. You run sqlc to generate Go code including iris-API & pg-orm 
1. You adjust application code based on generated struct, iris-API and pg-orm generated functions

Seriously, it's that easy. You don't have to write any boilerplate SQL querying
code ever again.

## Getting Started
Okay, enough hype, let's see it in action.

First you pass the following SQL schema file to `sqlc generate`:

```sql
CREATE TABLE authors (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  bio  text
);
```

And then in your application code you'll get generated delivery API, struct, Usecase and Database Repository:

Make sure you have adjust sqlc.yaml file before execute sqlc 
```yml
version: "1"
packages:
  - name: "db"
    path: "output" <--- output folder location
    project_path: "github.com/my-account/my-repo" <--- your golang project folder to generate import path
    queries: "./schema/query.sql" <--- ignore this one, just use dummy query string 
    schema: "./schema/author_schema.sql" <-- your target psotgre table schema 
    engine: "postgresql"
    emit_json_tags: true
    emit_form_tags: true
    emit_prepared_queries: true
    emit_interface: false
    emit_exact_table_names: false
    emit_empty_slices: false
    generate_model: true <--- set true to generate struct based on table schema
    generate_delivery: true <--- set true to generate iris-code delivery API based on table schema
    generate_repo: true <--- set true to generate pg-orm code based on table schema
    generate_usecase: true <--- set true to generate usecase code based on table schema
```

## Acknowledgements

sqlc was inspired by [PugSQL](https://pugsql.org/) and
[HugSQL](https://www.hugsql.org/) forked & modified from [SQLC](https://github.com/kyleconroy/sqlc)
