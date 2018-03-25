// Package vsmt provides simple migrations for PostgreSQL database but it don't supports rollback,
// squahs and complex features available in other migration packages.
package vsmt

import (
	"database/sql"
	"fmt"
)

// MigrationFunc for complex migrations. Execute any required queries inside this function. Don't commit transaction
// or rollback it, just return status through returned error value.
type MigrationFunc func(*sql.Tx) error

// Migrate list of migrations. Migration list can contain query strings or MigrationFuncs
func Migrate(tx *sql.Tx, scheme []string, migrations []interface{}) error {
	migrationNumber, err := getLastMigration(tx)
	if err != nil {
		return err
	}

	// empty database, so we just create schema and migration sequence
	if migrationNumber == 0 {
		return initSchemaAndMigration(tx, scheme)
	}

	if migrationNumber >= len(migrations) {
		// we don't have actual migrations
		return nil
	}

	// iterate over all migrations
	for i, migration := range migrations[migrationNumber:] {
		// execute migration string or function
		if err := execMigration(tx, migration); err != nil {
			return fmt.Errorf("exec migration #%v: %v", i, err)
		}
	}

	return nil
}

// execMigration query and updates current migration number
func execMigration(tx *sql.Tx, query interface{}) error {
	switch q := query.(type) {
	case string:
		// execute query string
		if _, err := tx.Exec(q); err != nil {
			return err
		}
	case MigrationFunc:
		// run migration functions
		if err := q(tx); err != nil {
			return err
		}
	default:
		return fmt.Errorf("migration item shuld be a query string or MigrationFunc type")
	}

	// update migration number
	q, err := tx.Query(`SELECT nextval('last_migration')`)
	if err != nil {
		return fmt.Errorf("increase migration number: %v", err)
	}

	return q.Close()
}

// getLastMigration if sequence is not set transaction will not affected. if migration number is 0 it means
// that database was empty, and we can create scheme from scratch. This method use PostgreSQL savepoint feature.
func getLastMigration(tx *sql.Tx) (migrationNumber int, err error) {
	if _, err := tx.Exec("SAVEPOINT check_last_migration"); err != nil {
		return migrationNumber, fmt.Errorf("start nested transaction: %v", err)
	}
	defer tx.Exec("RELEASE SAVEPOINT check_last_migration")

	if err := tx.QueryRow("SELECT last_value FROM last_migration").Scan(&migrationNumber); err != nil {
		// Here we rollback transaction to savepoint
		// This error indicates that we don't have last_migration object
		if _, err := tx.Exec("ROLLBACK TO SAVEPOINT check_last_migration"); err != nil {
			return migrationNumber, fmt.Errorf("rollback nested transaction: %v", err)
		}
	}
	return
}

// initSchemaAndMigration will creates all items from scheme list and 'last_migration'
func initSchemaAndMigration(tx *sql.Tx, scheme []string) error {
	// append sequence creation to scheme queries, and call nextval one time, to set it value 1
	scheme = append(scheme,
		"CREATE SEQUENCE last_migration",
		"SELECT nextval('last_migration')",
	)
	for i, query := range scheme {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("exec scheme query #%v: %v", i, err)
		}
	}

	// scheme is always actula, so no need exec migrations
	return nil
}
