package linkfixerbot

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"

	bolt "go.etcd.io/bbolt"
)

type Store interface {
	Put(string, Fixer) error
	Get(string) (Fixer, error)
	Delete(string) error
}

type BoltStore struct {
	db     *bolt.DB
	bucket []byte

	mu      sync.Mutex
	buf     bytes.Buffer
	encoder gob.Encoder
	decoder gob.Decoder
}

func NewBoltStore(db *bolt.DB, bucket []byte) (*BoltStore, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("could not ensure bucket %v exists: %v", string(bucket), err)
	}

	buf := bytes.Buffer{}
	return &BoltStore{
		db:      db,
		bucket:  bucket,
		encoder: *gob.NewEncoder(&buf),
		decoder: *gob.NewDecoder(&buf),
	}, nil
}

func (bs *BoltStore) encodeFixer(f Fixer) ([]byte, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.buf.Reset()
	err := bs.encoder.Encode(&f)
	if err != nil {
		return nil, err
	}

	return bs.buf.Bytes(), nil
}

func (bs *BoltStore) decodeFixer(b []byte) (Fixer, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.buf.Reset()
	var f Fixer
	err := bs.decoder.Decode(&f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (bs *BoltStore) Put(domain string, f Fixer) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bs.bucket)
		if b == nil {
			return fmt.Errorf("could not find bucket %v", bs.bucket)
		}

		fEncoded, err := bs.encodeFixer(f)
		if err != nil {
			return err
		}

		return b.Put([]byte(domain), fEncoded)
	})
}

func (bs *BoltStore) Get(domain string) (Fixer, error) {
	var res Fixer
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bs.bucket)
		if b == nil {
			return fmt.Errorf("could not find bucket %v", bs.bucket)
		}

		fEncoded := b.Get([]byte(domain))
		if fEncoded == nil {
			return nil
		}
		f, err := bs.decodeFixer(fEncoded)
		if err != nil {
			return err
		}

		res = f
		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
