package database

import (
	"sort"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	lastID := 0
	if len(dbStructure.Chirps) > 0 {
		lastChirp := dbStructure.Chirps[len(dbStructure.Chirps)]
		lastID = lastChirp.ID
	}
	newChirp := Chirp{
		Body:     body,
		ID:       lastID + 1,
		AuthorId: authorId,
	}

	dbStructure.Chirps[newChirp.ID] = newChirp
	err = db.writeDB(dbStructure)

	if err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}
	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, val := range dbStructure.Chirps {
		chirps = append(chirps, val)
	}
	sort.Slice(chirps, func(i, j int) bool {
		a, b := chirps[i], chirps[j]
		return a.ID < b.ID
	})
	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	chirp, ok := dbStructure.Chirps[id]
	if !ok {
		return Chirp{}, ErrNotExist
	}
	return chirp, nil
}

func (db *DB) DeleteChirp(id int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	delete(dbStructure.Chirps, id)
	db.writeDB(dbStructure)
	return nil
}
