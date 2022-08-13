package database

import (
	"fmt"
	"os"

	bolt "go.etcd.io/bbolt"

	"github.com/bytixo/AmberV2/internal/logger"
)

type DB struct {
	DB *bolt.DB
}

var (
	Proxy *DB
)

func init() {

	db, err := Setup()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	Proxy = &DB{
		DB: db,
	}
	logger.Info("Loaded Database")
}

//IsBlackListed fetch the dabatase and look if a userID
//have already been dm'ed, if yes it'll return true
func (db *DB) IsBlackListed(userID string) bool {
	var s bool
	_ = db.DB.View(func(tx *bolt.Tx) error {
		res := tx.Bucket([]byte("DB")).
			Bucket([]byte("DMED")).
			Get([]byte(userID))

		if res == nil {
			s = false
			return nil
		}
		s = true
		return nil
	})
	return s
}

//BlacklistID will blacklist the provided userID
func (db *DB) BlacklistID(userID string) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).
			Bucket([]byte("DMED")).
			Put([]byte(userID), []byte("true"))

		if err != nil {
			return fmt.Errorf("could not insert the provided userID: %v", err)
		}
		return nil
	})
	return err
}

func (db *DB) GetBlacklisted() int {
	var i int
	_ = db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("DMED"))
		b.ForEach(func(k, v []byte) error {
			i++
			return nil
		})
		return nil
	})
	return i
}
func Setup() (*bolt.DB, error) {
	db, err := bolt.Open("dmed.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't open db, %v", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("couldn't create root: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte("DMED"))
		if err != nil {
			return fmt.Errorf("couldn't create \"DMED\" DB: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	return db, nil
}
