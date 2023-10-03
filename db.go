package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type ChirpNotFound struct{}

func (e *ChirpNotFound) Error() string {
	return "Chirp not found"
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDB()
	if err != nil {
		return &DB{path: ""}, err
	}
	return db, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	chirps, err := db.GetChirps()
	if err != nil {
		return Chirp{}, errors.New("could not create chirp")
	}
	lastID := 0
	if len(chirps) > 0 {
		lastChirp := chirps[len(chirps)-1]
		lastID = lastChirp.ID
	}
	newChirp := Chirp{
		Body: body,
		ID:   lastID + 1,
	}

	dbData, err := db.loadDB()

	if err != nil {
		return Chirp{}, err
	}

	dbData.Chirps[newChirp.ID] = newChirp
	err = db.writeDB(dbData)

	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return make([]Chirp, 0), errors.New("could not get chirps")
	}
	chirps := make([]Chirp, 0)
	for _, val := range data.Chirps {
		chirps = append(chirps, val)
	}
	sort.Slice(chirps, func(i, j int) bool {
		a, b := chirps[i], chirps[j]
		return a.ID < b.ID
	})
	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, errors.New("could not get chirps")
	}
	chirp, ok := data.Chirps[id]
	if !ok {
		return Chirp{}, &ChirpNotFound{}
	}
	return chirp, nil
}

func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		structure := DBStructure{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}
		data, err := json.Marshal(structure)
		if err != nil {
			return err
		}
		os.WriteFile(db.path, data, 0777)
		return nil
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	structure := DBStructure{}
	err := db.ensureDB()
	if err != nil {
		return structure, err
	}

	file, err := os.ReadFile(db.path)
	if err != nil {
		return structure, err
	}

	err = json.Unmarshal(file, &structure)
	if err != nil {
		return structure, errors.New("could not decode database")
	}

	return structure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return errors.New("could not encode database")
	}
	err = os.WriteFile(db.path, data, fs.FileMode(os.O_RDWR))
	return err
}

func (db *DB) GetUsers() ([]User, error) {
	data, err := db.loadDB()
	if err != nil {
		return make([]User, 0), errors.New("could not get users")
	}
	users := make([]User, 0)
	for _, val := range data.Users {
		users = append(users, val)
	}
	sort.Slice(users, func(i, j int) bool {
		a, b := users[i], users[j]
		return a.ID < b.ID
	})
	return users, nil
}

func (db *DB) CreateUser(email string) (User, error) {
	users, err := db.GetUsers()
	if err != nil {
		return User{}, errors.New("could not create user")
	}
	lastID := 0
	if len(users) > 0 {
		lastUser := users[len(users)-1]
		lastID = lastUser.ID
	}
	newUser := User{
		Email: email,
		ID:    lastID + 1,
	}

	dbData, err := db.loadDB()

	if err != nil {
		return User{}, err
	}

	dbData.Users[newUser.ID] = newUser
	err = db.writeDB(dbData)

	if err != nil {
		return User{}, err
	}

	return newUser, nil
}
