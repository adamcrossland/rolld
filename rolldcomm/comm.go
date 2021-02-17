package rolldcomm

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/adamcrossland/rolld/roller"
	"github.com/gorilla/websocket"
)

// ConnectionInfo comprises information about an individual participant
// in a Session
type ConnectionInfo struct {
	ID         string
	Name       string
	Connection *websocket.Conn
}

// CommSession is all of the information about a Session
type CommSession struct {
	ID          string
	Connections map[string]*ConnectionInfo
	Commands    chan string
	members     string
}

// NewCommSession create a CommSession object that encapsulates all of the
// information that is unique to and required for a Session.
func NewCommSession(id string) *CommSession {
	newSess := new(CommSession)
	newSess.ID = id
	newSess.Connections = make(map[string]*ConnectionInfo)
	newSess.Commands = make(chan string, 100)

	go sharedProcessor(newSess)

	return newSess
}

func (session CommSession) broadcastMessage(message string) {
	messageAsBytes := []byte(message)

	for _, eachConn := range session.Connections {
		eachConn.Connection.WriteMessage(websocket.TextMessage, messageAsBytes)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// AddConnection adds a participant information to the Session's CommSession once
// they are ready to switch to WebSockets and begin communication.
func (session CommSession) AddConnection(id string, name string, w http.ResponseWriter, r *http.Request) {
	newConn := new(ConnectionInfo)
	newConn.ID = id
	newConn.Name = name
	var upgradeErr error
	newConn.Connection, upgradeErr = upgrader.Upgrade(w, r, nil)

	if upgradeErr != nil {
		log.Fatalf("Error upgrading connection to websocket: %v\n", upgradeErr)
	}

	session.Connections[id] = newConn

	session.Commands <- newConn.sendCommand([]string{"add", name})

	go func() {
		stillTicking := true

		for stillTicking {
			messType, message, err := newConn.Connection.ReadMessage()

			if err != nil {
				stillTicking = false
				delete(session.Connections, id)

				if strings.HasPrefix(err.Error(), "websocket: close") {
					// Client has closed the connection
					session.Commands <- fmt.Sprintf("%s quit %s", id, name)
					continue
				} else if strings.Contains(err.Error(), "connection timed out") {
					session.Commands <- fmt.Sprintf("%s timeout %s", id, name)
					continue
				} else {
					session.Commands <- fmt.Sprintf("%s dropped %s", id, name)
					log.Printf("error while reading websocket: %v", err)
					continue
				}
			}

			convertedMessage := string(message)
			convertedMessage = strings.ToLower(convertedMessage)

			messageParts := strings.Split(convertedMessage, " ")

			switch messageParts[0] {
			case "hello":
				newConn.Connection.WriteMessage(messType, []byte("Hello from rolld."))
			case "quit":
				newConn.Connection.WriteMessage(messType, []byte("bye"))
				newConn.Connection.Close()
				delete(session.Connections, id)
				session.Commands <- newConn.sendCommand([]string{"quit", name})
				stillTicking = false
			case "add":
				// This is not an allowed command. Just eat it.

			default:
				// All other messages are handled by the shared command
				// processor.
				session.Commands <- newConn.sendCommand(messageParts)
			}
		}
	}()
}

func (conn ConnectionInfo) sendCommand(commandParts []string) string {
	// Prepend the connection ID
	allParts := append([]string{conn.ID}, commandParts...)
	return strings.Join(allParts, " ")
}

func sharedProcessor(session *CommSession) {
	stillTicking := true

	log.Printf("Shared processor started for session %s\n", session.ID)

	for stillTicking {
		nextMessage := <-session.Commands
		log.Printf("Received command: %s\n", nextMessage)
		commandParts := strings.Split(nextMessage, " ")

		issuer := commandParts[0]
		command := commandParts[1]
		data := ""
		if len(commandParts) >= 3 {
			data = strings.Join(commandParts[2:], " ")
		}

		switch command {
		case "roll":
			spec, specErr := roller.Parse(data)
			if specErr != nil {
				errMessage := fmt.Sprintf("die roll format error: %v", specErr)
				session.Connections[issuer].Connection.WriteMessage(websocket.TextMessage, []byte(errMessage))
			} else {

				rollResults := roller.DoRolls(*spec)
				rollMessage := fmt.Sprintf("%s rolled %s: ", session.Connections[issuer].Name, data)
				for rollI := 0; rollI < rollResults.Count; rollI++ {
					rollMessage += fmt.Sprintf("%d ", rollResults.Rolls[rollI].Total)
				}

				session.broadcastMessage(rollMessage)
			}
		case "quit":
			session.members = ""
			byeMessage := fmt.Sprintf("%s has quit", data)
			session.broadcastMessage(byeMessage)
			session.Commands <- "0 members" // Force a members update

		case "timeout":
			session.members = ""
			byeMessage := fmt.Sprintf("%s was dropped for inactivity", data)
			session.broadcastMessage(byeMessage)
			session.Commands <- "0 members"

		case "dropped":
			session.members = ""
			byeMessage := fmt.Sprintf("%s dropped mysteriously", data)
			session.broadcastMessage(byeMessage)
			session.Commands <- "0 members"

		case "add":
			session.members = ""
			joinMessage := fmt.Sprintf("%s has joined.", data)
			session.broadcastMessage(joinMessage)
			session.Commands <- "0 members" // Force a members update

		case "members":
			if session.members == "" {
				session.buildMembersList()
			}

			session.broadcastMessage(fmt.Sprintf("members\n%s", session.members))
		default:
			errMessage := fmt.Sprintf("command not understood: %s", command)
			errMessageAsBytes := []byte(errMessage)
			session.Connections[issuer].Connection.WriteMessage(websocket.TextMessage, errMessageAsBytes)
		}
	}
}

func (session *CommSession) buildMembersList() {
	var tempList strings.Builder

	for _, v := range session.Connections {
		tempList.WriteString(fmt.Sprintf("%s\n", v.Name))
	}

	session.members = tempList.String()
}
