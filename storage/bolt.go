package storage

import (
	"github.com/boltdb/bolt"
	"time"
)

var database = []byte("nsearch")

const MODE = 0600

type BoltDB struct {
	db *bolt.DB
}

func NewBolt(path string) (Storage, error) {
	db, err := bolt.Open(path, MODE, &bolt.Options{Timeout: 3600 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(database)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &BoltDB {
		db : db,
	}, nil
}

func (bdb *BoltDB) AddData(k, v []byte) error {
	return bdb.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(database).Put(k, v)
	})
}

func (bdb *BoltDB) GetData(k []byte) (v []byte, err error) {
	err = bdb.db.View(func(tx *bolt.Tx) error {
		v = tx.Bucket(database).Get(k)
		return nil
	})
	return
}

func (bdb *BoltDB) DelData(k []byte) error {
	return bdb.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(database).Delete(k)
	})
}

func (bdb *BoltDB) ForEach(fn func(k, v []byte) error) error {
	return bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := fn(k, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bdb *BoltDB) Close() error {
	return bdb.db.Close()
}