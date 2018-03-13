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

	if _, err := db.Exec("DROP TABLE IF EXISTS table1"); err != nil {
		t.Error(err)
	}

	if _, err := db.Exec("DROP TABLE IF EXISTS table2"); err != nil {
		t.Error(err)
	}

	scheme := []string{
		`CREATE TABLE table1 (id serial, name text, name2 text)`,
		`CREATE TABLE table2 (id serial, name text)`,
	}

	migrations := []interface{}{
		`ALTER TABLE table1 ADD COLUMN name1 text`,
	}

	if err := Migrate(db, scheme, migrations); err != nil {
		t.Errorf("migrate schema: %v", err)
		return
	}

	t.Run("first insert", func(t *testing.T) {
		if _, err := db.Exec("INSERT INTO table2(name, name2) VALUES ($1, $2)", "name1", "name2"); err != nil {
			t.Errorf("can't insert data: %v", err)
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
