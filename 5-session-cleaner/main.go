//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"
)

const ttlSeconds = 5

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions              map[string]Session
	sessionExpirations    map[string]time.Time
	expirationChecks      map[string][]string
	expirationCheckTicker *time.Ticker
	mutex                 sync.RWMutex
}

// Session stores the session's data
type Session struct {
	Data map[string]interface{}
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions:              make(map[string]Session),
		sessionExpirations:    make(map[string]time.Time),
		expirationChecks:      make(map[string][]string),
		expirationCheckTicker: time.NewTicker(time.Second),
	}

	go m.removeExpiredSessionsWorker()

	return m
}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessions[sessionID] = Session{
		Data: make(map[string]interface{}),
	}

	m.updateSessionExpiration(sessionID)

	return sessionID, nil
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.mutex.RLock()
	_, ok := m.sessions[sessionID]
	if !ok {
		m.mutex.RUnlock()
		return ErrSessionNotFound
	}
	m.mutex.RUnlock()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Hint: you should renew expiry of the session here
	m.sessions[sessionID] = Session{
		Data: data,
	}

	m.updateSessionExpiration(sessionID)

	return nil
}

func (m *SessionManager) updateSessionExpiration(sessionID string) {
	expireAt := time.Now().Add(ttlSeconds * time.Second)
	m.sessionExpirations[sessionID] = expireAt
	key := strconv.Itoa(expireAt.Second())
	m.expirationChecks[key] = append(m.expirationChecks[key], sessionID)
}

func (m *SessionManager) removeExpiredSessionsWorker() {
	for range m.expirationCheckTicker.C {
		now := time.Now()
		key := strconv.Itoa(now.Second())

		m.mutex.RLock()
		expiredSessionIDs := []string{}
		for _, sID := range m.expirationChecks[key] {
			if exp, ok := m.sessionExpirations[sID]; ok {
				exp = exp.Add(-10 * time.Millisecond)
				if exp.Before(now) {
					expiredSessionIDs = append(expiredSessionIDs, sID)
				}

			}
		}
		m.mutex.RUnlock()

		m.mutex.Lock()
		for _, sID := range expiredSessionIDs {
			delete(m.sessions, sID)
			delete(m.sessionExpirations, sID)
		}
		delete(m.expirationChecks, key)
		m.mutex.Unlock()
	}
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}
