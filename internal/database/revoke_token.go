package database

import "time"

func (db *DB) GetTokenIsRevoked(token string) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}
	if _, ok := dbStructure.RevokedTokens[token]; ok {
		return true, nil
	}
	return false, nil
}

func (db *DB) AddRevokedToken(token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	dbStructure.RevokedTokens[token] = time.Now()
	db.writeDB(dbStructure)
	return nil
}
