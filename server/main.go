package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/adamcrossland/rolld/manageddb"
	"github.com/adamcrossland/rolld/models"
	"github.com/adamcrossland/rolld/rolldcomm"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var model *models.RolldModel
var sessions map[string]*rolldcomm.CommSession

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
	db := manageddb.NewManagedDB(dbFilename, "sqlite3", databaseMigrations)
	model = models.NewModel(db)

	r := mux.NewRouter()
	r.HandleFunc("/start/{connCount}", start)
	r.HandleFunc("/connect/{session}/{name}", connect)
	r.HandleFunc("/messages/{session}/{connection}", messages)
	r.HandleFunc("/hello", hello)
	r.HandleFunc("/client", client)
	r.HandleFunc("/robots.txt", robots)
	r.HandleFunc("/", client)
	http.Handle("/", r)

	servingAddress := os.Getenv("ROLLD_SERVER_ADDRESS")
	if servingAddress == "" {
		panic("environment variable ROLLD_SERVER_ADDRESS must be set")
	}
	fmt.Printf("Listening on %s\n", servingAddress)

	certPath := os.Getenv("ROLLD_SERVER_CERTPATH")
	if certPath == "" {
		panic("enviornment variable ROLLD_SERVER_CERTPATH must be set")
	}
	keyPath := os.Getenv("ROLLD_SERVER_KEYPATH")
	if keyPath == "" {
		panic("enviornment variable ROLLD_SERVER_KEYPATH must be set")
	}

	go http.ListenAndServe(":80", http.HandlerFunc(redirect))

	httpErr := http.ListenAndServeTLS(servingAddress, certPath, keyPath, nil)
	if httpErr != nil {
		log.Fatalf("error starting web server: %v\n", httpErr)
	}
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

// Initiate a new Rolld session. Returns a SessionID that must be included
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

	if sessions[sessionID] == nil {
		sessions[sessionID] = rolldcomm.NewCommSession(sessionID)
	}

	sessions[sessionID].AddConnection(connectionID, requestedConnection.Name, w, r)
}

// Send an acknowledgement message to let the client know that they are talking
// to a Rolld server.
func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "rolld ack")
}

// Serve the default web client file. By default, the file is cached to reduce
// disk I/O. This behavior can be over-ridden by starting the server with a
// parameter of --no-cached. The use case for this is doing development work on
// the client; changes can be tested without having to stop and re-start the
// server each time.
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

func robots(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "User-agent: *\nDisallow: /\n")
}
