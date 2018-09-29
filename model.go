package main

import (
	"fmt"
	"time"
)

type RolldModel struct {
	db ManagedDB
}

type Session struct {
	model *RolldModel
	ID string
	ConnectionCount int
	Created int64
}

func NewModel(db ManagedDB) *RolldModel {
	model = new(RolldModel)
	model.db = db

	return model
}

func (model *RolldModel) NewSession(connectionCount int) Session {
	stored := false
	timestamp := time.Now().Unix()

	var newSession Session

	for !stored {
		newSessionID = GetRandomID()
		_, err := model.DB.Exec("insert into sessions (id, connections, created) values (?,?,?)",
			newSessionID, connectionCount, timestamp)
		if err == nil {
			stored = true
			newSession.ID = newSessionID
			newSession.ConnectionCount = connectionCount
			newSession.Created = timestamp
			newSession.model = model
		}
	}

	return newSessionID
}

func (model *RolldModel) GetSession(sessionID string) (Session, error) {
	row, err := model.db.DB.QueryRow("select * from sessions where id = ?", sessionID)
	if err != nil {
		return nil, err
	}

	var foundSession Session
	foundSession.model = model
	row.Scan(&foundSession.Created)
	row.Scan(&foundSession.ConnectionCount)
	row.Scan(&foundSession.Created)

	return foundSession, nil
}

func (session Session) NameTaken(name string) bool {
	found := false
	row, err := session.model.db.DB.QueryRow("select count(1) from connections where session = ? and name = ?")
	if err = nil {
		var count int
		row.Scan(&count)
		if count > 0 {
			found = true
		}
	}

	return found
}

func (session Session) AddConnection(name string) error {
	if session.NameTaken(name) {
		return error("that name is already taken")
	}


}
