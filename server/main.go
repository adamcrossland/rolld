package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/adamcrossland/rolld/manageddb"
	"github.com/adamcrossland/rolld/models"
	"github.com/gorilla/mux"
)

var model *models.RolldModel

func main() {
	dbFilename := os.Getenv("ROLLD_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable ROLLD_DATABASE_FILE must be set")
	}
	db := manageddb.NewManagedDB(dbFilename, "sqlite3")
	model = models.NewModel(db)

	r := mux.NewRouter()
	r.HandleFunc("/start/{connCount}", start)
	r.HandleFunc("/connect/{session}/{name}", connect)
	http.Handle("/", r)

	servingAddress := os.Getenv("ROLLD_SERVER_ADDRESS")
	if servingAddress == "" {
		panic("environment variable ROLLD_SERVER_ADDRESS must be set")
	}
	fmt.Printf("Listening on %s\n", servingAddress)

	http.ListenAndServe(servingAddress, nil)
}

// Initiate a new rolld session. Returns a SessionID that must be included
// in all subsequent API calls.
func start(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestedSessionCount, parseErr := strconv.Atoi(vars["connCount"])
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
	if sessionID == "" {
		http.Error(w, "session must be provided", 400)
		return
	}

	requestedName := vars["name"]
	if requestedName == "" {
		http.Error(w, "name must be provided", 400)
		return
	}

	session, sesserr := model.GetSession(sessionID)
	if sesserr != nil {
		http.Error(w, "Session does not exist", 400)
		return
	}

	if session.NameTaken(requestedName) {
		http.Error(w, "The name is already taken", 400)
		return
	}

	conn, newConnErr := session.AddConnection(requestedName)
	if newConnErr != nil {
		http.Error(w, "Could not add connection. Try again later.", 500)
		return
	}

	fmt.Fprintf(w, "%s", conn.ID)

	return
}
