package database

import (
	"errors"
	"log"
)

type User struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password,omitempty"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (db *DB) CreateUser(email, password string) (User, error) {
	if _, err := db.GetUserByEmail(email); !errors.Is(err, ErrNotExist) {
		return User{}, ErrAlreadyExist
	}

	dbStructure, err := db.loadDB()
	if err != nil {
		log.Printf("LOAD DB ERROR, %v", err)
		return User{}, err
	}

	lastID := 0
	if len(dbStructure.Users) > 0 {
		lastUser := dbStructure.Users[len(dbStructure.Users)]
		lastID = lastUser.ID
	}
	newUser := User{
		Email:       email,
		Password:    password,
		ID:          lastID + 1,
		IsChirpyRed: false,
	}
	dbStructure.Users[newUser.ID] = newUser

	err = db.writeDB(dbStructure)
	if err != nil {
		log.Printf("WRITE DB ERROR, %v", err)
		return User{}, err
	}

	return newUser, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotExist
}

func (db *DB) GetUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	return user, nil
}

func (db *DB) UpdateUser(id int, email, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	user.Email = email
	user.Password = password
	dbStructure.Users[id] = user
	db.writeDB(dbStructure)

	return user, nil
}

func (db *DB) UpgradeUser(id int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return ErrNotExist
	}

	user.IsChirpyRed = true
	dbStructure.Users[id] = user
	db.writeDB(dbStructure)

	return nil
}
