package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type RolldModel struct {
	db *ManagedDB
}

type Session struct {
	model           *RolldModel
	ID              string
	ConnectionCount int
	Created         int64
}

type Connection struct {
	model     *RolldModel
	ID        string
	SessionID string
	Name      string
	Created   int64
}

func NewModel(db *ManagedDB) *RolldModel {
	model = new(RolldModel)
	model.db = db

	return model
}

func (model *RolldModel) NewSession(connectionCount int) Session {
	stored := false
	timestamp := time.Now().Unix()

	var newSession Session

	for !stored {
		newSessionID := GetRandomID()
		_, err := model.db.DB.Exec("insert into sessions (id, connections, created) values (?,?,?)",
			newSessionID, connectionCount, timestamp)
		if err == nil {
			stored = true
			newSession.ID = newSessionID
			newSession.ConnectionCount = connectionCount
			newSession.Created = timestamp
			newSession.model = model
		}
	}

	return newSession
}

func (model *RolldModel) GetSession(sessionID string) (*Session, error) {
	stmt, err := model.db.DB.Prepare("select id, connections, created from sessions where id = ?")
	if err != nil {
		fmt.Printf("error preparing session query: %v", err)
		return nil, nil
	}
	defer stmt.Close()

	fmt.Printf("Retrieving session %s\n", sessionID)
	var sessID string
	var count int
	var created int64
	rows, queryErr := stmt.Query(sessionID)

	var foundSession *Session

	if queryErr == nil || queryErr != sql.ErrNoRows {
		defer rows.Close()
		if rows.Next() {
			foundSession = new(Session)
			rows.Scan(&sessID, &count, &created)

			foundSession.model = model
			foundSession.ID = sessID
			foundSession.ConnectionCount = count
			foundSession.Created = created

			fmt.Printf("sessID = %s\n", sessID)
			fmt.Printf("count = %d\n", count)
			fmt.Printf("created = %d\n", created)
		} else {
			return nil, errors.New("Query GetSession returned 0 rows.")
		}
	} else {
		return nil, queryErr
	}

	return foundSession, nil
}

func (session Session) NameTaken(name string) bool {
	found := false
	row := session.model.db.DB.QueryRow("select count(1) from connections where session = ? and name = ?", session.ID, name)
	var count int
	row.Scan(&count)
	if count > 0 {
		found = true
	}

	return found
}

func (session Session) AddConnection(name string) (*Connection, error) {
	if session.NameTaken(name) {
		return nil, errors.New("that name is already taken")
	}

	saved := false
	var newConnection *Connection

	for !saved {
		newConnectionID := GetRandomID()
		timestamp := time.Now().Unix()
		_, err := session.model.db.DB.Exec("insert into connections (id, session, name, created) values (?, ?, ?, ?)", newConnectionID, session.ID, name, timestamp)
		if err != nil {
			return nil, err
		}

		saved = true

		newConnection = new(Connection)
		newConnection.ID = newConnectionID
		newConnection.SessionID = session.ID
		newConnection.Name = name
		newConnection.Created = timestamp
		newConnection.model = session.model
	}

	return newConnection, nil
}
