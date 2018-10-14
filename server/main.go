package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/adamcrossland/rolld/manageddb"
	"github.com/adamcrossland/rolld/models"
	"github.com/adamcrossland/rolld/rolldcomm"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var model *models.RolldModel
var sessions map[string]*rolldcomm.CommSession
var sessionLock *sync.Mutex
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var cachedClient []byte
var dontCache bool

func main() {
	argsWithoutProg := os.Args[1:]
	for i := 0; i < len(argsWithoutProg); i++ {
		switch strings.ToLower(argsWithoutProg[i]) {
		case "--no-cache":
			dontCache = true
		}
	}

	sessions = make(map[string]*rolldcomm.CommSession)

	dbFilename := os.Getenv("ROLLD_DATABASE_FILE")
	if dbFilename == "" {
		panic("environment variable ROLLD_DATABASE_FILE must be set")
	}
	db := manageddb.NewManagedDB(dbFilename, "sqlite3")
	model = models.NewModel(db)

	r := mux.NewRouter()
	r.HandleFunc("/start/{connCount}", start)
	r.HandleFunc("/connect/{session}/{name}", connect)
	r.HandleFunc("/messages/{session}/{connection}", messages)
	r.HandleFunc("/hello", hello)
	r.HandleFunc("/client", client)
	r.HandleFunc("/", client)
	http.Handle("/", r)

	servingAddress := os.Getenv("ROLLD_SERVER_ADDRESS")
	if servingAddress == "" {
		panic("environment variable ROLLD_SERVER_ADDRESS must be set")
	}
	fmt.Printf("Listening on %s\n", servingAddress)

	sessionLock = new(sync.Mutex)

	certPath := os.Getenv("ROLLD_SERVER_CERTPATH")
	if certPath == "" {
		panic("enviornment variable ROLLD_SERVER_CERTPATH must be set")
	}
	keyPath := os.Getenv("ROLLD_SERVER_KEYPATH")
	if keyPath == "" {
		panic("enviornment variable ROLLD_SERVER_KEYPATH must be set")
	}

	httpErr := http.ListenAndServeTLS(servingAddress, certPath, keyPath, nil)
	if httpErr != nil {
		log.Fatalf("error starting web server: %v\n", httpErr)
	}
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

// Claim one of the slots in a session
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

// Begin to send messages to the server that are processed and
// potentially communicated to other subscribers
func messages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	sessionID := vars["session"]
	if sessionID == "" {
		http.Error(w, "session must be provided", 400)
		return
	}

	requestedSession, reqSessErr := model.GetSession(sessionID)
	if reqSessErr != nil {
		http.Error(w, "could not find session in database", 500)
		return
	}

	connectionID := vars["connection"]
	if connectionID == "" {
		http.Error(w, "connection must be provided", 400)
		return
	}

	requestedConnection, reqConnErr := requestedSession.GetConnection(connectionID)
	if reqConnErr != nil {
		http.Error(w, "could not find connection in database", 500)
		return
	}

	sessionLock.Lock()
	if sessions[sessionID] == nil {
		sessions[sessionID] = rolldcomm.NewCommSession(sessionID)
	}
	sessionLock.Unlock()

	sessions[sessionID].AddConnection(connectionID, requestedConnection.Name, w, r)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "rolld ack")
}

func client(w http.ResponseWriter, r *http.Request) {
	// First, make sure that we have tyhe client loaded. We keep a cached copy
	// since it is not large and needn't be read from disk each time.
	if len(cachedClient) == 0 || dontCache {
		var clientLoadError error
		cachedClient, clientLoadError = ioutil.ReadFile("./rolld-client.html")
		if clientLoadError != nil {
			panic("unable to load client file")
		}
	}

	fmt.Fprintf(w, "%s", cachedClient)
}
