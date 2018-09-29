package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

var model RolldModel

func main() {
	dbFilename := os.Getenv("ROLLD_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable ROLLD_DATABASE_FILE must be set")
	}
	db := NewManagedDB(dbFilename, "sqlite3")
	model = NewModel(db)

	r := mux.NewRouter()
	r.HandleFunc("/start/{connCount}", start)
	r.HandleFunc("/connect/{session}/{name}", connect)
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

	newSession := model.NewSession(requestedSessionCount)
	fmt.Fprintf(w, "%s", newSession.ID)

	return
}

func connect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	sessionID := vars["session"]
	if sessionID == nil || sessionID == "" {
		http.Error(w, "session must be provided", 400)
		return
	}

	requestedName := vars["name"]
	if requestedName == nil || requestedName == "" {
		http.Error(w, "name must be provided", 400)
		return
	}

	session, sesserr := model.GetSession(sessionID)
	if sesserr != nil {
		http.Error(w, "Session does not exist", 400)
	}

}
