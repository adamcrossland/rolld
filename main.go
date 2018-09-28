package main

import (
	"os"
)

var db *ManagedDB

func main() {
	//http.HandleFunc("/start", start)
	dbFilename := os.Getenv("ROLLD_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable ROLLD_DATABASE_FILE must be set")
	}
	db = NewManagedDB(dbFilename, "sqlite3")
}

//func start(w http.ResponseWriter, r *http.Request) {

//}
