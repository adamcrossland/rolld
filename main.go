package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var db *ManagedDB

func main() {
	dbFilename := os.Getenv("ROLLD_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable ROLLD_DATABASE_FILE must be set")
	}
	db = NewManagedDB(dbFilename, "sqlite3")

	r := mux.NewRouter()
	r.HandleFunc("/start/{connCount}", start)
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

// Initiate a new rolld session. Returns a SessionID that must be included
// in all subsequent API calls.
func start(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedSessionCount, parseErr := strconv.ParseUint(vars["connCount"], 10, 8)
	if parseErr != nil {
		http.Error(w, "connCount could not be understood as an unsigned, 8-bit integer", 400)
		return
	}

	stored := false
	timestamp := time.Now().Unix()

	for !stored {
		newSessionID := GetRandomID()
		_, err := db.DB.Exec("insert into sessions (id, connections, created) values (?,?,?)",
			newSessionID, requestedSessionCount, timestamp)
		if err == nil {
			fmt.Fprintf(w, "%s", newSessionID)
			break
		}
	}

	return
}
