package manageddb

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Information pertaining to the database that is being managed. This should
// probably all be opaque to the user.
type ManagedDB struct {
	DB               *sql.DB
	dbLock           chan int
	migrations       map[int]dbMigration
	currentMigration int
}

// Create an initialize a new ManagedDB with the given file path and
// datatbase driver.
func NewManagedDB(dbPath string, driver string) *ManagedDB {
	newDB := new(ManagedDB)
	var err error

	newDB.DB, err = sql.Open(driver, dbPath)
	if err != nil {
		panic(fmt.Sprintf("err opening database: %v", err))
	}

	newDB.dbLock = make(chan int)

	// Figure out what the current migration is
	newDB.currentMigration = newDB.getCurrentMigration()
	log.Printf("current migration level: %d", newDB.currentMigration)
	newDB.migrations = map[int]dbMigration{
		1: dbMigration{up: migration1up, down: migration1down},
		2: dbMigration{up: migration2up, down: migration2down},
	}

	newDB.databaseMigrate(-1)
	return newDB
}

func (mdb ManagedDB) getCurrentMigration() int {
	var current int

	rows, err := mdb.DB.Query("select migration from db_metadata")
	if err == nil {
		defer rows.Close()
		rows.Next()
		if err = rows.Scan(&current); err != nil {
			log.Fatalf("Error getting current migration from db: %v", err)
			panic(err)
		}
	}

	return current
}

func (mdb ManagedDB) setCurrentMigration(level int) {
	_, err := mdb.DB.Exec("update db_metadata set migration = ?", level)
	if err != nil {
		log.Fatalf("unable to update database migration level: %v", err)
		panic(err)
	}
}

func (mdb ManagedDB) databaseMigrate(toMigration int) {
	// If desired migration level is -1, it means go to the latest
	// migration.
	if toMigration == -1 {
		toMigration = len(mdb.migrations)
	}

	var dbErr error

	if mdb.currentMigration > toMigration {
		// Migrating down.
		for mdb.currentMigration > toMigration {
			mdb.currentMigration--
			dbErr = mdb.migrations[mdb.currentMigration].down(mdb.DB)
			if dbErr != nil {
				panic(fmt.Sprintf("db migration %d down failed: %v", mdb.currentMigration, dbErr))
				mdb.currentMigration++
			} else {
				mdb.setCurrentMigration(mdb.currentMigration)
			}
		}
	} else if mdb.currentMigration < toMigration {
		// Migrating up.
		for mdb.currentMigration < toMigration {
			mdb.currentMigration++
			dbErr = mdb.migrations[mdb.currentMigration].up(mdb.DB)
			if dbErr != nil {
				panic(fmt.Sprintf("db migration %d up failed: %v", mdb.currentMigration, dbErr))
				mdb.currentMigration--
			} else {
				mdb.setCurrentMigration(mdb.currentMigration)
			}
		}
	}
}

type dbMigrationFunction func(db *sql.DB) error

type dbMigration struct {
	up   dbMigrationFunction
	down dbMigrationFunction
}

func migration1up(db *sql.DB) error {
	_, err := db.Exec("create table db_metadata (migration integer); insert into db_metadata (migration) values (0)")

	return err
}

func migration1down(db *sql.DB) error {
	_, err := db.Exec("drop table db_metadata")

	return err
}

func migration2up(db *sql.DB) error {
	_, err := db.Exec(`create table sessions (id text primary key, connections integer, created integer);
	create table connections (id text primary key, session text, name text, created integer)`)

	return err
}

func migration2down(db *sql.DB) error {
	_, err := db.Exec("drop table sessions")

	return err
}
