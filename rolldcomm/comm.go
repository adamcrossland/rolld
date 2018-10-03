package rolldcomm

import (
	"fmt"
	"log"
	"strings"

	"github.com/adamcrossland/rolld/roller"
	"github.com/gorilla/websocket"
)

type ConnectionInfo struct {
	ID         string
	Name       string
	Connection *websocket.Conn
}

type CommSession struct {
	ID          string
	Connections map[string]*ConnectionInfo
	Commands    chan string
}

func NewCommSession(id string) *CommSession {
	newSess := new(CommSession)
	newSess.ID = id
	newSess.Connections = make(map[string]*ConnectionInfo)
	newSess.Commands = make(chan string, 100)

	return newSess
}

func (session CommSession) BroadcastMessage(message string) {
	messageAsBytes := []byte(message)

	for _, eachConn := range session.Connections {
		eachConn.Connection.WriteMessage(websocket.TextMessage, messageAsBytes)
	}
}

func (session CommSession) AddConnection(id string, name string, conn *websocket.Conn) {
	newConn := new(ConnectionInfo)
	newConn.ID = id
	newConn.Name = name
	newConn.Connection = conn

	session.Connections[id] = newConn

	go func() {
		messType, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("error while reading websocket: %v", err)
		}

		convertedMessage := string(message)
		convertedMessage = strings.ToLower(convertedMessage)

		messageParts := strings.Split(convertedMessage, " ")

		switch messageParts[0] {
		case "hello":
			conn.WriteMessage(messType, []byte("ack"))
		case "quit":
			conn.WriteMessage(messType, []byte("bye"))
			conn.Close()
			session.Commands <- fmt.Sprintf("quit %s", name)
			delete(session.Connections, id)
		default:
			// All other messages are handled by the shared command
			// processor.
			session.Commands <- newConn.SendCommand(messageParts[1:])
		}
	}()
}

func (conn ConnectionInfo) SendCommand(commandParts []string) string {
	// Prepend the connection ID
	allParts := append([]string{conn.ID}, commandParts...)
	return strings.Join(allParts, " ")
}

func SharedProcessor(session *CommSession) {
	stillTicking := true

	for stillTicking {
		nextMessage := <-session.Commands
		commandParts := strings.Split(nextMessage, " ")

		issuer := commandParts[0]
		command := commandParts[1]
		data := commandParts[2]

		switch command {
		case "roll":
			spec, specErr := roller.Parse(data)
			if specErr != nil {
				errMessage := fmt.Sprintf("die roll format error: %v", specErr)
				session.Connections[issuer].Connection.WriteMessage(websocket.TextMessage, []byte(errMessage))
				return
			}

			rollResults := roller.DoRolls(*spec)
			rollMessage := fmt.Sprintf("%s rolled %s: %d", session.Connections[issuer].Name, data, rollResults.Count)

			session.BroadcastMessage(rollMessage)

		case "quit":
			byeMessage := fmt.Sprintf("%s has left", data)
			session.BroadcastMessage(byeMessage)

		default:
			errMessage := fmt.Sprintf("command not understood: %s", command)
			errMessageAsBytes := []byte(errMessage)
			session.Connections[issuer].Connection.WriteMessage(websocket.TextMessage, errMessageAsBytes)
		}
	}
}
