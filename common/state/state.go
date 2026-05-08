// Package state provides common state functionality
package state

import (
	"os"
	"path/filepath"
	"sync"

	bbolt "go.etcd.io/bbolt"
)

const stateBucket = "state"

type State struct {
	db *bbolt.DB
}

var (
	state *State
	mu    sync.Mutex
)

func Initialize(stateDir, dbName string) error {
	mu.Lock()
	defer mu.Unlock()

	if state != nil {
		return nil
	}

	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return err
	}

	db, err := bbolt.Open(filepath.Join(stateDir, dbName), 0600, nil)
	if err != nil {
		return err
	}

	state = &State{db: db}
	return nil
}

func Get() *State {
	mu.Lock()
	defer mu.Unlock()
	return state
}

func (s *State) PutByValue(key string, value string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(stateBucket))
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte(value))
	})
}

func (s *State) PutIfAbsent(key string, value string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(stateBucket))
		if err != nil {
			return err
		}

		if b.Get([]byte(key)) != nil {
			return nil
		}

		return b.Put([]byte(key), []byte(value))
	})
}

func (s *State) PutByFunc(key string, createValue func() (string, error)) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(stateBucket))
		if err != nil {
			return err
		}

		value, err := createValue()
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte(value))
	})
}

func (s *State) Exists(key string) (bool, error) {
	var exists bool

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(stateBucket))
		if b == nil {
			return nil
		}

		exists = b.Get([]byte(key)) != nil
		return nil
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *State) GetValue(key string) (string, error) {
	var result string

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(stateBucket))
		if b == nil {
			return nil
		}

		if v := b.Get([]byte(key)); v != nil {
			result = string(v)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return result, nil
}

func (s *State) GetOrComputeValue(
	key string,
	createValue func() (string, error),
) (string, error) {
	var result string

	err := s.db.Update(func(tx *bbolt.Tx) error {
		var err error
		result, err = getOrComputeValue(tx, key, createValue)
		return err
	})
	if err != nil {
		return "", err
	}

	return result, nil
}

func getOrComputeValue(
	tx *bbolt.Tx,
	key string,
	createValue func() (string, error),
) (string, error) {
	b, err := tx.CreateBucketIfNotExists([]byte(stateBucket))
	if err != nil {
		return "", err
	}

	// Already exists
	if v := b.Get([]byte(key)); v != nil {
		return string(v), nil
	}

	// Create new value
	v, err := createValue()
	if err != nil {
		return "", err
	}

	if err := b.Put([]byte(key), []byte(v)); err != nil {
		return "", err
	}

	return v, nil
}
