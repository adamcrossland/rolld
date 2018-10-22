package main

import (
	"database/sql"

	"github.com/adamcrossland/rolld/manageddb"
)

var databaseMigrations map[int]manageddb.DBMigration

func init() {
	databaseMigrations = map[int]manageddb.DBMigration{
		1: manageddb.DBMigration{Up: migration1up, Down: migration1down},
		2: manageddb.DBMigration{Up: migration2up, Down: migration2down},
	}
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
