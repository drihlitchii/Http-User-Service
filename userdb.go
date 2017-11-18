package main

import (
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
)

type DB struct {
	db *bolt.DB
}

func NewDB(path string) (*DB, error) {
	db, err := bolt.Open(path, 0644, nil)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

//User data
type User struct {
	ID    string
	Fname string `json:"fname"`
	Sname string `json:"sname"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	About string `json:"about"`
}

//AddUser ...
func (db *DB) AddUser(u User) error {

	tx, err := db.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	b, err := tx.CreateBucketIfNotExists([]byte("users"))
	if err != nil {
		return err
	}
	juser, err := json.Marshal(&u)
	if err != nil {
		return err
	}
	if err := b.Put([]byte(u.ID), juser); err != nil {
		return err
	}

	return tx.Commit()
}

//GetUser ...
func (db *DB) GetUser(id string) (*User, error) {
	tx, err := db.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	b := tx.Bucket([]byte("users"))
	juser := b.Get([]byte(id))
	if juser == nil {
		return nil, errors.New("user not found")
	}
	var user User
	err = json.Unmarshal(juser, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) AllUsers() ([]User, error) {
	var users []User
	err := db.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("users"))

		err := b.ForEach(func(k, v []byte) error {
			var u User
			err := json.Unmarshal(v, &u)
			if err != nil {
				return err
			}
			users = append(users, u)
			return nil
		})
		return err
	})
	return users, err
}
