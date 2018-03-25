package vsmt

import (
	"fmt"
	"os"
	"testing"

	"database/sql"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func TestMigrations(t *testing.T) {
	db := getDB(t)
	defer db.Close()

	scheme := []string{
		`CREATE TABLE table1 (id serial, name text)`,
		`CREATE TABLE table2 (id serial, name text, name2 text)`,
	}

	migrations := []interface{}{
		`ALTER TABLE table1 ADD COLUMN name text`,
	}

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("create transaction: %v", err)
		return
	}
	defer tx.Rollback()

	if err := Migrate(tx, scheme, migrations); err != nil {
		t.Errorf("migrate schema: %v", err)
		return
	}

	t.Run("first insert", func(t *testing.T) {
		if _, err := tx.Exec("INSERT INTO table2(name, name2) VALUES ($1, $2)", "name1", "name2"); err != nil {
			t.Errorf("can't insert data: %v", err)
		}
	})

	t.Run("call migration second time", func(t *testing.T) {
		if err := Migrate(tx, scheme, migrations); err != nil {
			t.Errorf("migrate schema second time: %v", err)
		}
	})

	t.Run("add one more migration", func(t *testing.T) {
		scheme[0] = `CREATE TABLE table1 (id serial, name text, name2 text)`
		migrations = append(migrations, "ALTER TABLE table1 ADD COLUMN name2 text")

		if err := Migrate(tx, scheme, migrations); err != nil {
			t.Errorf("migrate schema second time: %v", err)
		}
	})

	t.Run("second insert", func(t *testing.T) {
		if _, err := tx.Exec("INSERT INTO table1(name, name2) VALUES ($1, $2);", "name1", "name2"); err != nil {
			t.Errorf("can't insert data: %+v", err)
		}
	})

	t.Run("add migration through function", func(t *testing.T) {
		migrations = append(migrations,
			MigrationFunc(func(tx *sql.Tx) error {
				q, err := tx.Query("SELECT 1")
				if err != nil {
					return err
				}

				return q.Close()
			}))
		if err := Migrate(tx, scheme, migrations); err != nil {
			t.Errorf("migrate schema second time: %v", err)
		}
	})
}

func getDB(t *testing.T) *sql.DB {
	cs := fmt.Sprintf(
		"user=%v dbname=%v password=%v host=%v port=%v",
		os.Getenv("VSMT_USER"),
		os.Getenv("VSMT_NAME"),
		os.Getenv("VSMT_PASS"),
		os.Getenv("VSMT_HOST"),
		os.Getenv("VSMT_PORT"),
	)

	db, err := sql.Open("postgres", cs)
	if err != nil {
		t.Fatalf("open connection to test database: %v", err)
		return nil
	}
	return db
}

func init() {
	godotenv.Load(".env")
}
