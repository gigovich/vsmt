[![Go Report Card](https://goreportcard.com/badge/github.com/gigovich/vsmt)](https://goreportcard.com/report/github.com/gigovich/vsmt)

# Very Simple Migration Tool
This package don't supports rollback, squahs and complex features available in other migration packages.
It's handle only PostgreSQL and uses sequences feature to store migration number.

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
In example below, you can add any number of query strings, which will be executed only once. Instead strings, you can use `MigrationFunc` functions, which receives transaction and
returns error. Inside this functions you can do any manipulations with DB, but don't rollback or commit transactions, it will be done in caller function and depends on your returned error value.
```go
package model

import (
	"github.com/gigovich/ethshot/service/db"
	"github.com/gigovich/vsmt"
)

var migrations = []interface{}{
	`CREATE TABLE job (
		id serial,
		created_at timestamp with time zone,
		updated_at timestamp with time zone,
		deleted_at timestamp with time zone,
		screenshot_id integer,
		processed bool,
		progress integer,
		error text,
		state text
	)`,
	`ALTER TABLE job ADD COLUMN address text`,
	`ALTER TABLE job DROP COLUMN state`,
}

func init() {
	if err := vsmt.Migrate(db.Default.DB, migrations); err != nil {
		panic("do migrations: " + err.Error())
	}
}
```
