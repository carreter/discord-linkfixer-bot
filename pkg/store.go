package linkfixerbot

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/charmbracelet/log"
	bolt "go.etcd.io/bbolt"
)

type Store interface {
	Put(guildID string, domain string, f Fixer) error
	Get(guildID string, domain string) (Fixer, error)
	Delete(guildID string, domain string) error
	List(guildID string) (map[string]Fixer, error)
}

type FixerList []struct {
	Domain string
	Fixer  Fixer
}

type BoltStore struct {
	db *bolt.DB
}

func NewBoltStore(db *bolt.DB) *BoltStore {
	return &BoltStore{
		db: db,
	}
}

func (bs *BoltStore) encodeFixer(f Fixer) ([]byte, error) {
	b := bytes.Buffer{}
	err := gob.NewEncoder(&b).Encode(&f)
	if err != nil {
		log.Error("could not encode fixer", "fixer", f, "err", err)
		return nil, err
	}

	log.Debug("encoded fixer", "encoded", b.Bytes())

	return b.Bytes(), nil
}

func (bs *BoltStore) decodeFixer(b []byte) (Fixer, error) {
	var f Fixer
	err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&f)
	if err != nil {
		log.Error("could not decode fixer", "err", err)
		return nil, err
	}

	return f, nil
}

func (bs *BoltStore) Delete(guildID string, domain string) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(guildID))
		if b == nil {
			return fmt.Errorf("could not find bucket %v", guildID)
		}

		return b.Delete([]byte(domain))
	})
}

func (bs *BoltStore) Put(guildID string, domain string, f Fixer) error {
	return bs.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(guildID))
		if err != nil {
			return err
		}

		fEncoded, err := bs.encodeFixer(f)
		if err != nil {
			return err
		}

		err = b.Put([]byte(domain), fEncoded)
		if err != nil {
			return err
		}

		log.Info("created fixer", "guildID", guildID, "domain", domain, "fixer", f)
		return nil
	})
}

func (bs *BoltStore) Get(guildID string, domain string) (Fixer, error) {
	var res Fixer
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(guildID))
		if b == nil {
			return fmt.Errorf("could not find bucket %v", guildID)
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

func (bs *BoltStore) List(guildID string) (map[string]Fixer, error) {
	res := map[string]Fixer{}
	err := bs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(guildID))
		if b == nil {
			return fmt.Errorf("could not find bucket %v", guildID)
		}

		return b.ForEach(func(domain, fEncoded []byte) error {
			f, err := bs.decodeFixer(fEncoded)
			if err != nil {
				return err
			}

			res[string(domain)] = f
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
