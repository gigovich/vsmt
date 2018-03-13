package vsmt

import (
	"database/sql"
	"fmt"
)

// MigrationFunc for complex migrations. Execute any required queries inside this function. Don't commit transaction
// or rollback it, just return status through returned error value.
type MigrationFunc func(*sql.Tx) error

// Migrate list of migrations. Migration list can contain query strings or MigrationFuncs
func Migrate(db *sql.DB, scheme []string, migrations []interface{}) error {
	// prepend migrations list with initial query which creates migration sequence
	migrations = append([]interface{}{"CREATE SEQUENCE last_migration START WITH 1"}, migrations...)

	var migrationNumber int
	if err := db.QueryRow("SELECT last_value FROM last_migration").Scan(&migrationNumber); err != nil {
		// Here we can log error, but in most cases this error indicates that we don't have last_migration object
		// log.Println("get last migration number: %v", err)
	}

	// inital migration, create scheme
	if migrationNumber == 0 {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %v", err)
		}

		for i, query := range scheme {
			if _, err := tx.Exec(query); err != nil {
				tx.Rollback()
				return fmt.Errorf("exec scheme query #%v: %v", i, err)
			}
		}

		_, err = tx.Exec(fmt.Sprintf("CREATE SEQUENCE last_migration START WITH %d", len(migrations)))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("set correct last migration number: %v", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit scheme: %v", err)
		}

		// scheme is always actula, so no need exec migrations
		return nil
	}

	if migrationNumber >= len(migrations)-1 {
		// we don't have actual migrations
		return nil
	}

	// iterate over all migrations
	for i, migration := range migrations[migrationNumber:] {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %v", err)
		}

		// execute migration string or function
		if err := exec(tx, migration); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec migration #%v: %v", i, err)
		}

		// update migration number
		if _, err := tx.Exec(`SELECT nextval('last_migration')`); err != nil {
			tx.Rollback()
			return fmt.Errorf("increase migration number: %v", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration: %v", err)
		}
	}

	return nil
}

// exec query and updates current migration number
func exec(tx *sql.Tx, query interface{}) error {
	switch q := query.(type) {
	case string:
		// execute query string
		_, err := tx.Exec(q)
		return err
	case MigrationFunc:
		// run migration functions
		return q(tx)
	default:
		return fmt.Errorf("migration item shuld be a query string or MigrationFunc type")
	}
}
