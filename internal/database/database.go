package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]User         `json:"users"`
	RevokedTokens map[string]time.Time `json:"revoked_tokens"`
}

var ErrNotExist = errors.New("resource does not exist")
var ErrAlreadyExist = errors.New("resource already exists")

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) createDB() error {
	structure := DBStructure{
		Chirps:        make(map[int]Chirp),
		Users:         make(map[int]User),
		RevokedTokens: make(map[string]time.Time),
	}
	return db.writeDB(structure)
}

func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbStructure := DBStructure{}
	file, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(file, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	return os.WriteFile(db.path, data, 0600)
}

func (db *DB) ResetDB() error {
	err := os.Remove(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	db.ensureDB()
	return err
}
