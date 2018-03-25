[![Go Report Card](https://goreportcard.com/badge/github.com/gigovich/vsmt)](https://goreportcard.com/report/github.com/gigovich/vsmt)

# Very Simple Migration Tool

This package don't supports rollback, squahs and complex features available in other migration packages.
It handles only PostgreSQL and uses [SEQUENCE](https://www.postgresql.org/docs/8.3/static/sql-createsequence.html) feature to store migration number.

## Install

We recommend use [dep](https://github.com/golang/dep) to install package:
```bash
$ dep ensure -add github.com/gigovich/vsmt
```
or simple `go get`
```bash
$ go get github.com/gigovich/vsmt
```

## Usage

Migration function uses two arguments `scheme` and `migrations`. First argument should contain actual state of schema. Queries from this list always used
for database initialization. Second argument contains migration queries.

Don't forget update `scheme` when add queries to `migrations` argument. You can call migration function any time on empty/non empty databases and it will guaratiee you actual state of scheme.

Besides use plain queries inside `migrations` argument, you can use `MigrationFunc` which help you do complex queries where you need Golang code fox processing data or queries.

```go
package main

import (
	"database/sql"
	"fmt"

	"github.com/gigovich/vsmt"
	_ "github.com/lib/pq"
)

// schema should contain actual state of database, when you create migration comman you also should update
// this records
var schema = []string{
	`CREATE TABLE state (
		id serial PRIMARY KEY,
		name text,
		is_active bool
	)`,
	`CREATE TABLE job (
		id serial PRIMARY KEY,
		state_id int REFERENCES state(id),
		title text
	)`,
}

// migrations queries list, when you add query here don't forget modify schema according this query
var migrations = []interface{}{
	`ALTER TABLE state ADD COLUMN is_active bool`,
	`ALTER TABLE job DROP COLUMN created_at`,
	// migration can be not just SQL strings, but also callback functions. This is very usefull for datamigations
	// or other complex migrations where you can use golang code for some generation or processing.
	vsmt.MigrationFunc(func(tx *sql.Tx) error {
		for _, state := range []string{"Started", "In Proccess", "Finished"} {
			if _, err := tx.Exec("INSERT INTO state(name) VALUES($1)", state); err != nil {
				return err
			}
		}
		return nil
	}),
}

func main() {
	// create database connection
	db, err := sql.Open("postgres", "postgres://vsmt:vsmt@localhost/vsmt")
	if err != nil {
		fmt.Println("open database:", err)
		return
	}

	// migrations function expect open transaction and will exequte all queries inside this transaction
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("create migration tx:", err)
		return
	}

	// defer commit migration transaction
	defer func() {
		if err := tx.Commit(); err != nil {
			fmt.Println("commit transaction:", err)
			return
		}
		fmt.Println("database migrated")
	}()

	// run migration process, if this first time call schema queries will be exequted else required queries from
	// migrations argument
	if err := vsmt.Migrate(tx, schema, migrations); err != nil {
		fmt.Println("do migrations:", err)
	}

}
```

## Tests

For testing you should use real database connection.

1. Create database
2. Copy `.env_example` to `.evn`
3. Edit `.env` and replace `VSMT_*` variables with your values
4. Run testing `go test .`
